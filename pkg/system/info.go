package system

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
)

// Info represents system information
type Info struct {
	OS           string `json:"os"`
	Architecture string `json:"architecture"`
	CPUCores     int    `json:"cpu_cores"`
	TotalRAM     string `json:"total_ram"`
	RAMBytes     uint64 `json:"ram_bytes"`
	Platform     string `json:"platform"`
	GoVersion    string `json:"go_version"`
	Hostname     string `json:"hostname"`
}

// GetSystemInfo returns comprehensive system information
func GetSystemInfo() (*Info, error) {
	info := &Info{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		CPUCores:     runtime.NumCPU(),
		GoVersion:    runtime.Version(),
	}
	
	// Get hostname
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	} else {
		info.Hostname = "unknown"
	}
	
	// Get platform string
	info.Platform = getPlatformString(info.OS, info.Architecture)
	
	// Get memory information
	ramBytes, ramString := getMemoryInfo()
	info.RAMBytes = ramBytes
	info.TotalRAM = ramString
	
	return info, nil
}

// getPlatformString returns a user-friendly platform string
func getPlatformString(os, arch string) string {
	osName := os
	archName := arch
	
	// Convert OS names to user-friendly format
	switch strings.ToLower(os) {
	case "darwin":
		osName = "macOS"
	case "linux":
		osName = "Linux"
	case "windows":
		osName = "Windows"
	case "freebsd":
		osName = "FreeBSD"
	case "openbsd":
		osName = "OpenBSD"
	case "netbsd":
		osName = "NetBSD"
	}
	
	// Convert architecture names to user-friendly format
	switch strings.ToLower(arch) {
	case "amd64":
		archName = "x64"
	case "arm64":
		archName = "ARM64"
	case "386":
		archName = "x86"
	case "arm":
		archName = "ARM"
	case "mips":
		archName = "MIPS"
	case "mips64":
		archName = "MIPS64"
	case "ppc64":
		archName = "PowerPC64"
	case "ppc64le":
		archName = "PowerPC64LE"
	case "s390x":
		archName = "IBM Z"
	}
	
	return fmt.Sprintf("%s %s", osName, archName)
}

// getMemoryInfo returns memory information in bytes and formatted string
func getMemoryInfo() (uint64, string) {
	switch runtime.GOOS {
	case "linux":
		return getLinuxMemory()
	case "darwin":
		return getDarwinMemory()
	case "windows":
		return getWindowsMemory()
	case "freebsd", "openbsd", "netbsd":
		return getBSDMemory()
	default:
		// Fallback estimation
		return getEstimatedMemory()
	}
}

// getLinuxMemory reads memory information from /proc/meminfo
func getLinuxMemory() (uint64, string) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return getEstimatedMemory()
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				if kb, err := strconv.ParseUint(fields[1], 10, 64); err == nil {
					bytes := kb * 1024
					return bytes, formatMemorySize(bytes)
				}
			}
		}
	}
	
	return getEstimatedMemory()
}

// getDarwinMemory uses sysctl to get memory information on macOS
func getDarwinMemory() (uint64, string) {
	cmd := exec.Command("sysctl", "-n", "hw.memsize")
	output, err := cmd.Output()
	if err != nil {
		return getEstimatedMemory()
	}
	
	size, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return getEstimatedMemory()
	}
	
	return size, formatMemorySize(size)
}

// getWindowsMemory uses PowerShell to get memory information on Windows
func getWindowsMemory() (uint64, string) {
	// Try using wmic first
	cmd := exec.Command("wmic", "computersystem", "get", "TotalPhysicalMemory", "/value")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "TotalPhysicalMemory=") {
				sizeStr := strings.TrimPrefix(line, "TotalPhysicalMemory=")
				if size, err := strconv.ParseUint(sizeStr, 10, 64); err == nil {
					return size, formatMemorySize(size)
				}
			}
		}
	}
	
	// Fallback to PowerShell
	cmd = exec.Command("powershell", "-Command", "(Get-CimInstance Win32_PhysicalMemory | Measure-Object -Property capacity -Sum).sum")
	output, err = cmd.Output()
	if err != nil {
		return getEstimatedMemory()
	}
	
	size, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return getEstimatedMemory()
	}
	
	return size, formatMemorySize(size)
}

// getBSDMemory uses sysctl to get memory information on BSD systems
func getBSDMemory() (uint64, string) {
	cmd := exec.Command("sysctl", "-n", "hw.physmem")
	output, err := cmd.Output()
	if err != nil {
		return getEstimatedMemory()
	}
	
	size, err := strconv.ParseUint(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return getEstimatedMemory()
	}
	
	return size, formatMemorySize(size)
}

// getEstimatedMemory provides a fallback memory estimation
func getEstimatedMemory() (uint64, string) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Conservative estimate: multiply system memory by reasonable factor
	estimated := m.Sys * 8
	minMem := uint64(2 * 1024 * 1024 * 1024) // 2GB minimum
	
	if estimated < minMem {
		estimated = minMem
	}
	
	// Cap at reasonable maximum
	maxMem := uint64(128 * 1024 * 1024 * 1024) // 128GB maximum estimation
	if estimated > maxMem {
		estimated = maxMem
	}
	
	return estimated, formatMemorySize(estimated)
}

// formatMemorySize formats bytes into human-readable string
func formatMemorySize(bytes uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
		TB = GB * 1024
	)
	
	switch {
	case bytes >= TB:
		return fmt.Sprintf("%.1f TB", float64(bytes)/TB)
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/GB)
	case bytes >= MB:
		return fmt.Sprintf("%.0f MB", float64(bytes)/MB)
	case bytes >= KB:
		return fmt.Sprintf("%.0f KB", float64(bytes)/KB)
	default:
		return fmt.Sprintf("%d bytes", bytes)
	}
}

// GetCPUInfo returns detailed CPU information
func GetCPUInfo() map[string]interface{} {
	info := map[string]interface{}{
		"cores":        runtime.NumCPU(),
		"architecture": runtime.GOARCH,
		"go_max_procs": runtime.GOMAXPROCS(0),
	}
	
	// Try to get more detailed CPU info on different platforms
	switch runtime.GOOS {
	case "linux":
		if cpuInfo := getLinuxCPUInfo(); len(cpuInfo) > 0 {
			for k, v := range cpuInfo {
				info[k] = v
			}
		}
	case "darwin":
		if cpuInfo := getDarwinCPUInfo(); len(cpuInfo) > 0 {
			for k, v := range cpuInfo {
				info[k] = v
			}
		}
	case "windows":
		if cpuInfo := getWindowsCPUInfo(); len(cpuInfo) > 0 {
			for k, v := range cpuInfo {
				info[k] = v
			}
		}
	}
	
	return info
}

// getLinuxCPUInfo reads CPU information from /proc/cpuinfo
func getLinuxCPUInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	data, err := os.ReadFile("/proc/cpuinfo")
	if err != nil {
		return info
	}
	
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				
				switch key {
				case "model name":
					info["model"] = value
					return info // We got what we need
				case "cpu family":
					info["family"] = value
				case "vendor_id":
					info["vendor"] = value
				}
			}
		}
	}
	
	return info
}

// getDarwinCPUInfo uses sysctl to get CPU information on macOS
func getDarwinCPUInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	// Get CPU brand
	if cmd := exec.Command("sysctl", "-n", "machdep.cpu.brand_string"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			info["model"] = strings.TrimSpace(string(output))
		}
	}
	
	// Get CPU vendor
	if cmd := exec.Command("sysctl", "-n", "machdep.cpu.vendor"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			info["vendor"] = strings.TrimSpace(string(output))
		}
	}
	
	return info
}

// getWindowsCPUInfo uses wmic to get CPU information on Windows
func getWindowsCPUInfo() map[string]interface{} {
	info := make(map[string]interface{})
	
	cmd := exec.Command("wmic", "cpu", "get", "name", "/value")
	output, err := cmd.Output()
	if err != nil {
		return info
	}
	
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Name=") {
			info["model"] = strings.TrimPrefix(line, "Name=")
			break
		}
	}
	
	return info
}