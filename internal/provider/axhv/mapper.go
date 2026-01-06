package axhv

import (
	"fmt"
	"strconv"
	"strings"

	"aexon/internal/provider/axhv/pb"
	"aexon/internal/types"
	"aexon/internal/utils"
)

// MapCreateRequestV2 maps values directly without parsing from strings.
// This is the preferred method when the frontend sends numeric values.
func MapCreateRequestV2(name string, image string, vcpu int, memoryMiB int, diskGB int, bandwidthMbps int, ip string, gateway string, ports map[string]string, password string) (*pb.CreateVmRequest, error) {
	// Apply defaults
	if vcpu <= 0 {
		vcpu = 1
	}
	if memoryMiB <= 0 {
		memoryMiB = 512
	}
	if diskGB <= 0 {
		diskGB = 10
	}

	// Default password if not provided
	if password == "" {
		password = "root"
	}

	// Map Image to Paths
	kernelPath, rootfsPath, err := mapImageToPaths(image)
	if err != nil {
		return nil, err
	}

	// Parse Ports from limits map
	portMap := make(map[uint32]uint32)
	if portsStr, ok := ports["ports"]; ok {
		rules := strings.Split(portsStr, ",")
		for _, rule := range rules {
			parts := strings.Split(rule, ":")
			if len(parts) == 2 {
				hostPort, _ := strconv.Atoi(parts[0])
				guestPort, _ := strconv.Atoi(parts[1])
				if hostPort > 0 && guestPort > 0 {
					portMap[uint32(hostPort)] = uint32(guestPort)
				}
			}
		}
	}

	pbReq := &pb.CreateVmRequest{
		Id:                 name,
		Vcpu:               uint32(vcpu),
		MemoryMib:          uint32(memoryMiB),
		DiskSizeGb:         uint32(diskGB),
		BandwidthLimitMbps: uint32(bandwidthMbps),
		GuestIp:            ip,
		GuestGateway:       gateway,
		KernelPath:         kernelPath,
		RootfsPath:         rootfsPath,
		PortMapTcp:         portMap,
		RootPassword:       password,
	}

	// Note: Free tier limits are NOT applied here - caller can enforce if needed
	return pbReq, nil
}

// MapCreateRequest maps the internal CreateInstanceRequest to the protobuf CreateVmRequest.
// It also enforces Free Tier limitations.
func MapCreateRequest(req types.Instance, ip string, gateway string) (*pb.CreateVmRequest, error) {

	// Parse Limits
	cpu := utils.ParseCpuCores(req.Limits["cpu"])
	if cpu == 0 {
		cpu = 1
	}

	ram := utils.ParseMemoryToMB(req.Limits["memory"])
	if ram == 0 {
		ram = 512
	}

	disk := uint32(10) // Default 10GB if not specified
	if val, ok := req.Limits["disk"]; ok {
		// Simplified parsing, assuming GB integer for now or implementing parser
		d, _ := strconv.Atoi(val)
		if d > 0 {
			disk = uint32(d)
		}
	}

	// Parse Ports
	portMap := make(map[uint32]uint32)
	if val, ok := req.Limits["ports"]; ok {
		// Input: "2202:22,8080:80" (hostPort:guestPort)
		rules := strings.Split(val, ",")
		for _, rule := range rules {
			parts := strings.Split(rule, ":")
			if len(parts) == 2 {
				hostPort, _ := strconv.Atoi(parts[0])
				guestPort, _ := strconv.Atoi(parts[1])

				if hostPort > 0 && guestPort > 0 {
					portMap[uint32(hostPort)] = uint32(guestPort)
				}
			}
		}
	}

	// Map Image to Paths
	kernelPath, rootfsPath, err := mapImageToPaths(req.Image)
	if err != nil {
		return nil, err
	}

	pbReq := &pb.CreateVmRequest{
		Id:           req.Name,
		Vcpu:         uint32(cpu),
		MemoryMib:    uint32(ram),
		DiskSizeGb:   disk,
		GuestIp:      ip,
		GuestGateway: gateway,
		KernelPath:   kernelPath,
		RootfsPath:   rootfsPath,
		PortMapTcp:   portMap,
	}

	// Enforce Free Tier Limits (Hardcoded enforcement for now as requested)
	// In a real scenario, we might check req.Plan or User context.
	// Assuming all creations via this path are subject to these rules for the task context "Free Tier Enforcement".

	applyFreeTierLimits(pbReq)

	return pbReq, nil
}

func mapImageToPaths(imageName string) (string, string, error) {
	// Base directories for AxHV
	kernelDir := "/var/lib/axhv/kernels"
	imagesDir := "/var/lib/axhv/images"

	// Use the default kernel
	kernelPath := fmt.Sprintf("%s/vmlinux-distro", kernelDir)

	// Normalize image name and map to rootfs
	switch {
	case strings.Contains(imageName, "ubuntu"):
		return kernelPath,
			fmt.Sprintf("%s/ubuntu-rootfs.ext4", imagesDir),
			nil
	case strings.Contains(imageName, "alpine"):
		// Alpine might use same kernel but different rootfs
		return kernelPath,
			fmt.Sprintf("%s/alpine-rootfs.ext4", imagesDir),
			nil
	default:
		// Default to ubuntu
		return kernelPath,
			fmt.Sprintf("%s/ubuntu-rootfs.ext4", imagesDir),
			nil
	}
}

func applyFreeTierLimits(req *pb.CreateVmRequest) {
	// Bandwidth: 0 = unlimited (no traffic shaping)
	// Removed: req.BandwidthLimitMbps = 10

	// 2. Port Limits
	// As we don't have ports in the generic input yet (usually added later),
	// we initialize the maps to empty or filtered if they were passed.
	// If the request had ports (e.g. from a rich request object), we would filter them here.
	// Since types.Instance doesn't strictly have a list of initial ports in its basic struct
	// (usually added via AddPort), we ensure the map is initialized to allow strict validation if we were to add them.

	// Example of restricting if we were populating from a source that had ports:
	limitTcp := 3
	limitUdp := 1

	if len(req.PortMapTcp) > limitTcp {
		newMap := make(map[uint32]uint32)
		i := 0
		for k, v := range req.PortMapTcp {
			if i >= limitTcp {
				break
			}
			newMap[k] = v
			i++
		}
		req.PortMapTcp = newMap
	}

	if len(req.PortMapUdp) > limitUdp {
		newMap := make(map[uint32]uint32)
		i := 0
		for k, v := range req.PortMapUdp {
			if i >= limitUdp {
				break
			}
			newMap[k] = v
			i++
		}
		req.PortMapUdp = newMap
	}
}
