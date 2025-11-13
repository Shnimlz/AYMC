package core

import (
	"testing"
)

func TestNewSystemMonitor(t *testing.T) {
	monitor := NewSystemMonitor()

	if monitor == nil {
		t.Fatal("Monitor es nil")
	}
}

func TestGetMetrics(t *testing.T) {
	monitor := NewSystemMonitor()
	metrics := monitor.GetMetrics()

	if metrics == nil {
		t.Fatal("Metrics es nil")
	}

	// Verificar que las métricas tienen valores razonables
	if metrics.CPUPercent < 0 || metrics.CPUPercent > 100 {
		t.Errorf("CPUPercent fuera de rango: %f", metrics.CPUPercent)
	}

	if metrics.MemoryPercent < 0 || metrics.MemoryPercent > 100 {
		t.Errorf("MemoryPercent fuera de rango: %f", metrics.MemoryPercent)
	}

	if metrics.DiskPercent < 0 || metrics.DiskPercent > 100 {
		t.Errorf("DiskPercent fuera de rango: %f", metrics.DiskPercent)
	}

	if metrics.MemoryTotal == 0 {
		t.Error("MemoryTotal es 0")
	}

	if metrics.DiskTotal == 0 {
		t.Error("DiskTotal es 0")
	}
}

func TestGetOpenPorts(t *testing.T) {
	monitor := NewSystemMonitor()
	ports, err := monitor.GetOpenPorts()

	if err != nil {
		t.Errorf("Error obteniendo puertos abiertos: %v", err)
	}

	// Los puertos pueden estar vacíos, pero no debería ser nil
	if ports == nil {
		t.Error("Ports es nil")
	}

	// Verificar que los puertos están en rango válido
	for _, port := range ports {
		if port < 0 || port > 65535 {
			t.Errorf("Puerto fuera de rango: %d", port)
		}
	}
}
