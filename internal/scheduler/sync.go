package scheduler

import (
	"database/sql"
	"log"
	"strings"

	"aexon/internal/db"
	"aexon/internal/provider/lxc"
	"aexon/internal/types"
)

// RunStartupSync synchronizes instances from the LXD provider to the database.
func RunStartupSync(dbConn *sql.DB, lxd *lxc.InstanceService) {
	log.Println("[Sync] Starting LXD to DB synchronization...")

	lxdInstances, err := lxd.ListInstances()
	if err != nil {
		log.Printf("[Sync] ERROR: Failed to list instances from LXD: %v", err)
		return
	}

	for _, lxdInstance := range lxdInstances {
		dbInstance, err := db.GetInstance(lxdInstance.Name)
		if err != nil {
			if err == sql.ErrNoRows {
				// Instance does not exist in DB, let's import it.
				log.Printf("[Sync] Importing new instance '%s' from LXD to database...", lxdInstance.Name)

				newInstance := &types.Instance{
					Name:            lxdInstance.Name,
					Image:           lxdInstance.Config["volatile.base_image"],
					Limits:          lxdInstance.Config,
					Type:            lxdInstance.Type,
					BackupSchedule:  "@daily", // Default value
					BackupRetention: 7,        // Default value
					BackupEnabled:   false,    // Default value
				}

				if err := db.CreateInstance(newInstance); err != nil {
					log.Printf("[Sync] ERROR: Failed to import instance '%s': %v", lxdInstance.Name, err)
				} else {
					log.Printf("[Sync] Imported instance '%s' successfully.", lxdInstance.Name)
				}
			} else {
				log.Printf("[Sync] ERROR: Failed to query instance '%s' from DB: %v", lxdInstance.Name, err)
			}
		} else {
			// Instance exists in DB, update its status and IP addresses
			log.Printf("[Sync] Updating existing instance '%s' with LXD status...", lxdInstance.Name)

			// Update the instance in the database with current status and IP addresses from LXD
			instanceState, _, stateErr := lxd.GetInstanceState(lxdInstance.Name)
			if stateErr != nil {
				log.Printf("[Sync] Warning: Could not get state for instance '%s': %v", lxdInstance.Name, stateErr)
				// Still update the status from the list if we can't get the state
				dbInstance.Limits["status"] = strings.ToUpper(lxdInstance.Status)
			} else {
				// Extract IP addresses from the state
				var ipv4, ipv6 string

				if eth0, ok := instanceState.Network["eth0"]; ok {
					// Find IPv4 and IPv6 addresses
					for _, addr := range eth0.Addresses {
						if addr.Family == "inet" {
							ipv4 = addr.Address
						} else if addr.Family == "inet6" && ipv6 == "" {
							ipv6 = addr.Address
						}
					}
				}

				// Update the limits map with the IP addresses
				if ipv4 != "" {
					dbInstance.Limits["volatile.ipv4"] = ipv4
				} else {
					delete(dbInstance.Limits, "volatile.ipv4") // Remove if not available
				}
				if ipv6 != "" {
					dbInstance.Limits["volatile.ipv6"] = ipv6
				} else {
					delete(dbInstance.Limits, "volatile.ipv6") // Remove if not available
				}

				// Update status with normalized uppercase value
				dbInstance.Limits["status"] = strings.ToUpper(instanceState.Status)
			}

			// Update the instance in the database
			if err := db.UpdateInstanceStatusAndLimits(dbInstance.Name, dbInstance.Limits); err != nil {
				log.Printf("[Sync] ERROR: Failed to update instance '%s': %v", lxdInstance.Name, err)
			} else {
				log.Printf("[Sync] Updated instance '%s' with current status and IPs.", lxdInstance.Name)
			}
		}
	}
	log.Println("[Sync] Synchronization finished.")
}
