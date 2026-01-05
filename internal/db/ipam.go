package db

import (
	"context"
	"database/sql"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"time"
)

type Network struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CIDR      string    `json:"cidr"`
	Gateway   string    `json:"gateway"`
	DNS1      string    `json:"dns1"`
	VlanID    int       `json:"vlan_id"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
}

// AllocateIP finds a free IP across available networks using a "Smart Pool" strategy.
// It supports both pre-populated (legacy) and sparse (new) allocation models.
func (s *Service) AllocateIP(ctx context.Context, instanceName string) (string, error) {
	// 1. Determine Plan Type (Placeholder for now, default to Free/Private)
	// In the future, we can check user quota/plan here.
	isPro := false

	// 2. Fetch candidate networks
	networks, err := s.getAvailableNetworks(ctx, isPro)
	if err != nil {
		return "", fmt.Errorf("failed to fetch networks: %w", err)
	}

	// 3. Try allocation in each network
	for _, net := range networks {
		ip, err := s.tryAllocateInNetwork(ctx, net, instanceName)
		if err == nil {
			log.Printf("[IPAM] Allocated %s from network %s (%s)", ip, net.Name, net.CIDR)
			return ip, nil
		}
		// Log but continue to next network
		// log.Printf("[IPAM] Pool %s full or error: %v", net.Name, err)
	}

	return "", fmt.Errorf("no IP addresses available in any pool")
}

// AllocateInNetwork allocates an IP in a specific network pool.
func (s *Service) AllocateInNetwork(ctx context.Context, networkID string, instanceName string) (string, error) {
	var net Network
	query := `SELECT id, name, cidr, gateway, dns1, vlan_id, is_public FROM networks WHERE id = $1`
	err := s.QueryRowContext(ctx, query, networkID).Scan(&net.ID, &net.Name, &net.CIDR, &net.Gateway, &net.DNS1, &net.VlanID, &net.IsPublic)
	if err != nil {
		return "", fmt.Errorf("network not found: %w", err)
	}

	ip, err := s.tryAllocateInNetwork(ctx, net, instanceName)
	if err != nil {
		return "", fmt.Errorf("allocation failed in pool %s: %w", net.Name, err)
	}
	return ip, nil
}

func (s *Service) getAvailableNetworks(ctx context.Context, isPro bool) ([]Network, error) {
	query := `SELECT id, name, cidr, gateway, dns1, vlan_id, is_public FROM networks WHERE is_public = $1 ORDER BY created_at ASC`

	// Free plan gets Private (is_public=false). Pro logic handles both later.
	// For now, simple bool.

	rows, err := s.QueryContext(ctx, query, isPro)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var networks []Network
	for rows.Next() {
		var n Network
		if err := rows.Scan(&n.ID, &n.Name, &n.CIDR, &n.Gateway, &n.DNS1, &n.VlanID, &n.IsPublic); err != nil {
			return nil, err
		}
		networks = append(networks, n)
	}
	return networks, nil
}

type NetworkStats struct {
	Network
	TotalIPs     int     `json:"total_ips"`
	UsedIPs      int     `json:"used_ips"`
	UsagePercent float64 `json:"usage_percent"`
}

func (s *Service) GetNetworksWithStats(ctx context.Context) ([]NetworkStats, error) {
	// Fetch all networks
	query := `SELECT id, name, cidr, gateway, dns1, vlan_id, is_public, created_at FROM networks ORDER BY created_at ASC`
	rows, err := s.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []NetworkStats
	for rows.Next() {
		var n NetworkStats
		if err := rows.Scan(&n.ID, &n.Name, &n.CIDR, &n.Gateway, &n.DNS1, &n.VlanID, &n.IsPublic, &n.CreatedAt); err != nil {
			return nil, err
		}

		// Calculate Total IPs
		_, ipNet, _ := net.ParseCIDR(n.CIDR)
		if ipNet != nil {
			ones, _ := ipNet.Mask.Size()
			n.TotalIPs = 1 << (32 - ones)
			if n.TotalIPs > 2 {
				n.TotalIPs -= 3 // Network, Gateway, Broadcast
			}
		}

		// Count Used IPs
		countQuery := `SELECT COUNT(*) FROM ip_leases WHERE network_id = $1 AND instance_name IS NOT NULL`
		s.QueryRowContext(ctx, countQuery, n.ID).Scan(&n.UsedIPs)

		if n.TotalIPs > 0 {
			n.UsagePercent = (float64(n.UsedIPs) / float64(n.TotalIPs)) * 100
		}

		stats = append(stats, n)
	}

	return stats, nil
}

func (s *Service) CreateNetwork(ctx context.Context, n Network) error {
	query := `INSERT INTO networks (name, cidr, gateway, is_public) VALUES ($1, $2, $3, $4)`
	_, err := s.ExecContext(ctx, query, n.Name, n.CIDR, n.Gateway, n.IsPublic)
	return err
}

func (s *Service) tryAllocateInNetwork(ctx context.Context, netDef Network, instanceName string) (string, error) {
	// 1. Calculate Range
	startIP, endIP, err := CidrToRange(netDef.CIDR)
	if err != nil {
		return "", err
	}

	// DEBUG: Log the calculated range
	log.Printf("[IPAM-DEBUG] Network=%s CIDR=%s StartIP=%d EndIP=%d TotalIPs=%d", netDef.Name, netDef.CIDR, startIP, endIP, endIP-startIP)

	// Start searching from Start + 2 (Skipping Network .0 and Gateway .1)
	currentIP := startIP + 2

	// 2. Fetch ALL used IPs in this network (ignoring placeholders)
	query := `SELECT ip FROM ip_leases WHERE network_id = $1 AND instance_name IS NOT NULL`
	rows, err := s.QueryContext(ctx, query, netDef.ID)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	usedMap := make(map[string]bool)
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			return "", err
		}
		usedMap[ip] = true
	}

	log.Printf("[IPAM-DEBUG] UsedIPs in network: %d", len(usedMap))

	// 3. Find First Free IP
	attemptCount := 0
	for i := currentIP; i < endIP; i++ {
		ipStr := IntToIP(i)
		attemptCount++

		if !usedMap[ipStr] {
			// Found candidate! Try to reserve using Transaction for safety
			tx, err := s.BeginTx(ctx, nil)
			if err != nil {
				log.Printf("[IPAM-DEBUG] Failed to begin TX: %v", err)
				return "", err
			}

			// Check if row exists (in ANY network - ip is PK)
			var existsGlobal bool
			tx.QueryRowContext(ctx, "SELECT EXISTS(SELECT 1 FROM ip_leases WHERE ip = $1)", ipStr).Scan(&existsGlobal)

			if existsGlobal {
				// Row exists - try to claim it for THIS network
				res, err := tx.ExecContext(ctx,
					"UPDATE ip_leases SET instance_name = $1, allocated_at = $2, network_id = $3 WHERE ip = $4 AND instance_name IS NULL",
					instanceName, time.Now(), netDef.ID, ipStr)
				if err != nil {
					log.Printf("[IPAM-DEBUG] UPDATE failed for %s: %v", ipStr, err)
					tx.Rollback()
					continue
				}
				rowsAff, _ := res.RowsAffected()
				if rowsAff == 0 {
					log.Printf("[IPAM-DEBUG] UPDATE affected 0 rows for %s (already taken?)", ipStr)
					tx.Rollback()
					continue
				}
			} else {
				// Insert new lease
				_, err := tx.ExecContext(ctx,
					"INSERT INTO ip_leases (ip, instance_name, allocated_at, network_id) VALUES ($1, $2, $3, $4)",
					ipStr, instanceName, time.Now(), netDef.ID)
				if err != nil {
					log.Printf("[IPAM-DEBUG] INSERT failed for %s: %v", ipStr, err)
					tx.Rollback()
					continue
				}
			}

			if err := tx.Commit(); err != nil {
				log.Printf("[IPAM-DEBUG] COMMIT failed for %s: %v", ipStr, err)
				continue
			}

			log.Printf("[IPAM-DEBUG] SUCCESS: Allocated %s after %d attempts", ipStr, attemptCount)
			return ipStr, nil
		}
	}

	log.Printf("[IPAM-DEBUG] POOL_FULL after checking %d IPs", attemptCount)
	return "", fmt.Errorf("POOL_FULL")
}

// ReleaseIP frees the IP assigned to an instance.
func (s *Service) ReleaseIP(ctx context.Context, instanceName string) error {
	// We just clear the ownership. We keep the row (switch to Pre-populated mode basically)
	// Or we could Delete if we want to stay Sparse.
	// For "Hybrid" stability, keeping it NULL is fine and safer for logs.
	query := `
        UPDATE ip_leases 
        SET instance_name = NULL, allocated_at = NULL 
        WHERE instance_name = $1
    `

	_, err := s.ExecContext(ctx, query, instanceName)
	if err != nil {
		return fmt.Errorf("failed to release IP for instance %s: %w", instanceName, err)
	}

	return nil
}

// GetInstanceIP retrieves the IP assigned to an instance.
func (s *Service) GetInstanceIP(ctx context.Context, instanceName string) (string, error) {
	query := `SELECT ip FROM ip_leases WHERE instance_name = $1`

	var ip string
	err := s.QueryRowContext(ctx, query, instanceName).Scan(&ip)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", nil // Not found, return empty
		}
		return "", err
	}

	return ip, nil
}

// --- Extended Types for Admin UI ---

type IpLease struct {
	IP           string     `json:"ip_address"`
	InstanceName *string    `json:"instance_name"`
	AllocatedAt  *time.Time `json:"allocated_at"`
	Status       string     `json:"status"` // "allocated" or "reserved"
}

type NetworkDetails struct {
	Network
	Stats  NetworkStats `json:"stats"`
	Leases []IpLease    `json:"leases"`
}

// GetNetworkDetails fetches a specific network with its usage stats and full lease list.
func (s *Service) GetNetworkDetails(ctx context.Context, id string) (*NetworkDetails, error) {
	// 1. Fetch Network
	query := `SELECT id, name, cidr, gateway, dns1, vlan_id, is_public, created_at FROM networks WHERE id = $1`
	var n Network
	err := s.QueryRowContext(ctx, query, id).Scan(&n.ID, &n.Name, &n.CIDR, &n.Gateway, &n.DNS1, &n.VlanID, &n.IsPublic, &n.CreatedAt)
	if err != nil {
		return nil, err
	}

	details := &NetworkDetails{
		Network: n,
		Leases:  []IpLease{},
	}

	// 2. Calculate Stats (Total/Used)
	// (Reusing logic from GetNetworksWithStats basically, but for single ID)
	_, ipNet, _ := net.ParseCIDR(n.CIDR)
	if ipNet != nil {
		ones, _ := ipNet.Mask.Size()
		details.Stats.TotalIPs = 1 << (32 - ones)
		if details.Stats.TotalIPs > 2 {
			details.Stats.TotalIPs -= 3
		}
	}
	details.Stats.Network = n // Copy base info

	// 3. Fetch Leases
	// We want ALL leases for this network
	leasesQuery := `SELECT ip, instance_name, allocated_at FROM ip_leases WHERE network_id = $1 ORDER BY ip`
	rows, err := s.QueryContext(ctx, leasesQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	usedCount := 0
	for rows.Next() {
		var l IpLease
		var instName sql.NullString
		var allocAt sql.NullTime

		if err := rows.Scan(&l.IP, &instName, &allocAt); err != nil {
			return nil, err
		}

		if instName.Valid {
			l.InstanceName = &instName.String
			l.Status = "allocated"
			usedCount++
		} else {
			l.Status = "reserved" // Pre-allocated but not assigned to VM
		}

		if allocAt.Valid {
			l.AllocatedAt = &allocAt.Time
		}

		details.Leases = append(details.Leases, l)
	}

	details.Stats.UsedIPs = usedCount
	if details.Stats.TotalIPs > 0 {
		details.Stats.UsagePercent = (float64(usedCount) / float64(details.Stats.TotalIPs)) * 100
	}

	return details, nil
}

// DeleteNetwork removes a network pool. Fails if there are active allocations.
func (s *Service) DeleteNetwork(ctx context.Context, id string) error {
	// Check for active allocations
	var count int
	err := s.QueryRowContext(ctx, "SELECT COUNT(*) FROM ip_leases WHERE network_id = $1 AND instance_name IS NOT NULL", id).Scan(&count)
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("network has %d active IP allocations", count)
	}

	// Delete leases first (if cascading isn't set up or to be safe)
	// Actually schema migration 8 didn't specify ON DELETE CASCADE explicitly for the foreign key,
	// but usually we want to clean up.
	// Migration 8: FOREIGN KEY (network_id) REFERENCES networks(id)
	// Default is NO ACTION. So we MUST delete leases first.
	_, err = s.ExecContext(ctx, "DELETE FROM ip_leases WHERE network_id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to cleanup leases: %w", err)
	}

	// Delete network
	res, err := s.ExecContext(ctx, "DELETE FROM networks WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// --- Helpers ---

func CidrToRange(cidr string) (uint32, uint32, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return 0, 0, err
	}

	// Forçar conversão para 4 bytes (IPv4)
	ip4 := ip.To4()
	if ip4 == nil {
		return 0, 0, fmt.Errorf("IPv6 not supported in this pool")
	}

	// Get mask size (e.g. 24 for /24)
	ones, _ := ipnet.Mask.Size()

	// Recreate clean 32-bit mask
	mask := binary.BigEndian.Uint32(net.CIDRMask(ones, 32))
	start := binary.BigEndian.Uint32(ip4)

	// Calculate end by inverting mask
	end := (start & mask) | (mask ^ 0xffffffff)

	// DEBUG:
	// log.Printf("[IPAM-DEBUG] CIDR=%s Start=%d End=%d Total=%d", cidr, start, end, end-start)

	return start, end, nil
}

func IntToIP(nn uint32) string {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, nn)
	return ip.String()
}
