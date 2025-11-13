package core

import (
	"log"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

// SystemMonitor monitorea los recursos del sistema
type SystemMonitor struct {
	lastCheck time.Time
}

// SystemMetrics contiene las métricas del sistema
type SystemMetrics struct {
	Timestamp     time.Time         `json:"timestamp"`
	CPUPercent    float64           `json:"cpu_percent"`
	MemoryTotal   uint64            `json:"memory_total"`
	MemoryUsed    uint64            `json:"memory_used"`
	MemoryPercent float64           `json:"memory_percent"`
	DiskTotal     uint64            `json:"disk_total"`
	DiskUsed      uint64            `json:"disk_used"`
	DiskPercent   float64           `json:"disk_percent"`
	NetworkSent   uint64            `json:"network_sent"`
	NetworkRecv   uint64            `json:"network_recv"`
	Uptime        uint64            `json:"uptime"`
	ProcessCount  int               `json:"process_count"`
	Platform      string            `json:"platform"`
	PlatformVersion string          `json:"platform_version"`
	KernelVersion string            `json:"kernel_version"`
	OpenPorts     []int             `json:"open_ports"`
}

// NewSystemMonitor crea un nuevo monitor de sistema
func NewSystemMonitor() *SystemMonitor {
	return &SystemMonitor{
		lastCheck: time.Now(),
	}
}

// GetMetrics obtiene las métricas actuales del sistema
func (sm *SystemMonitor) GetMetrics() *SystemMetrics {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// CPU
	if cpuPercent, err := cpu.Percent(time.Second, false); err == nil && len(cpuPercent) > 0 {
		metrics.CPUPercent = cpuPercent[0]
	} else {
		log.Printf("[WARN] Error obteniendo uso de CPU: %v", err)
	}

	// Memoria
	if vmStat, err := mem.VirtualMemory(); err == nil {
		metrics.MemoryTotal = vmStat.Total
		metrics.MemoryUsed = vmStat.Used
		metrics.MemoryPercent = vmStat.UsedPercent
	} else {
		log.Printf("[WARN] Error obteniendo memoria: %v", err)
	}

	// Disco
	if diskStat, err := disk.Usage("/"); err == nil {
		metrics.DiskTotal = diskStat.Total
		metrics.DiskUsed = diskStat.Used
		metrics.DiskPercent = diskStat.UsedPercent
	} else {
		log.Printf("[WARN] Error obteniendo uso de disco: %v", err)
	}

	// Red
	if netStats, err := net.IOCounters(false); err == nil && len(netStats) > 0 {
		metrics.NetworkSent = netStats[0].BytesSent
		metrics.NetworkRecv = netStats[0].BytesRecv
	} else {
		log.Printf("[WARN] Error obteniendo estadísticas de red: %v", err)
	}

	// Información del host
	if hostInfo, err := host.Info(); err == nil {
		metrics.Uptime = hostInfo.Uptime
		metrics.Platform = hostInfo.Platform
		metrics.PlatformVersion = hostInfo.PlatformVersion
		metrics.KernelVersion = hostInfo.KernelVersion
		metrics.ProcessCount = int(hostInfo.Procs)
	} else {
		log.Printf("[WARN] Error obteniendo información del host: %v", err)
	}

	// Información de Go runtime
	metrics.Platform = runtime.GOOS

	sm.lastCheck = time.Now()

	return metrics
}

// GetCPUInfo obtiene información detallada de CPU
func (sm *SystemMonitor) GetCPUInfo() ([]cpu.InfoStat, error) {
	return cpu.Info()
}

// GetDiskPartitions obtiene las particiones de disco
func (sm *SystemMonitor) GetDiskPartitions() ([]disk.PartitionStat, error) {
	return disk.Partitions(false)
}

// GetNetworkInterfaces obtiene las interfaces de red
func (sm *SystemMonitor) GetNetworkInterfaces() ([]net.InterfaceStat, error) {
	return net.Interfaces()
}

// CheckJavaInstalled verifica si Java está instalado
func (sm *SystemMonitor) CheckJavaInstalled() (bool, string, error) {
	// TODO: Implementar verificación de Java
	return false, "", nil
}

// GetOpenPorts obtiene los puertos abiertos en el sistema
func (sm *SystemMonitor) GetOpenPorts() ([]int, error) {
	connections, err := net.Connections("tcp")
	if err != nil {
		return nil, err
	}

	portsMap := make(map[int]bool)
	for _, conn := range connections {
		if conn.Status == "LISTEN" {
			portsMap[int(conn.Laddr.Port)] = true
		}
	}

	ports := make([]int, 0, len(portsMap))
	for port := range portsMap {
		ports = append(ports, port)
	}

	return ports, nil
}
