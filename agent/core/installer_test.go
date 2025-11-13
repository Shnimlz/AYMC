package core

import (
	"os/exec"
	"runtime"
	"testing"
)

func TestNewJavaInstaller(t *testing.T) {
	installer, err := NewJavaInstaller("21")
	if err != nil {
		t.Fatalf("Error creando JavaInstaller: %v", err)
	}

	if installer == nil {
		t.Fatal("JavaInstaller es nil")
	}

	if installer.version != "21" {
		t.Errorf("Versión esperada '21', obtenida '%s'", installer.version)
	}

	if installer.osType == "" {
		t.Error("osType no fue detectado")
	}

	info := installer.GetInfo()
	if info["os"] == "" {
		t.Error("OS info vacía")
	}
}

func TestDetectSystem(t *testing.T) {
	installer, err := NewJavaInstaller("21")
	if err != nil {
		t.Fatalf("Error creando JavaInstaller: %v", err)
	}

	// Verificar que detectó correctamente el SO
	switch runtime.GOOS {
	case "linux":
		if installer.packageMgr == "" {
			t.Error("No se detectó el gestor de paquetes en Linux")
		}
		if installer.javaPackage == "" {
			t.Error("No se determinó el paquete de Java en Linux")
		}

	case "windows":
		if installer.packageMgr != "choco" {
			t.Errorf("Se esperaba 'choco' en Windows, obtenido '%s'", installer.packageMgr)
		}

	case "darwin":
		if installer.packageMgr != "brew" {
			t.Errorf("Se esperaba 'brew' en macOS, obtenido '%s'", installer.packageMgr)
		}
	}
}

func TestCheckInstalled(t *testing.T) {
	installer, err := NewJavaInstaller("21")
	if err != nil {
		t.Fatalf("Error creando JavaInstaller: %v", err)
	}

	installed, version, err := installer.CheckInstalled()
	if err != nil {
		t.Errorf("Error verificando instalación: %v", err)
	}

	// Si java está instalado, debe retornar versión
	if installed && version == "" {
		t.Error("Java reportado como instalado pero sin versión")
	}

	// Si no está instalado, verificar que realmente no esté
	if !installed {
		_, err := exec.LookPath("java")
		if err == nil {
			t.Error("Java está en PATH pero CheckInstalled retornó false")
		}
	}
}

func TestGetInfo(t *testing.T) {
	installer, err := NewJavaInstaller("17")
	if err != nil {
		t.Fatalf("Error creando JavaInstaller: %v", err)
	}

	info := installer.GetInfo()

	requiredFields := []string{"os", "package_mgr", "java_package", "version"}
	for _, field := range requiredFields {
		if _, exists := info[field]; !exists {
			t.Errorf("Campo requerido '%s' no existe en info", field)
		}
	}

	if info["version"] != "17" {
		t.Errorf("Versión esperada '17', obtenida '%s'", info["version"])
	}
}

func TestLinuxDistroDetection(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("Test solo para Linux")
	}

	installer, err := NewJavaInstaller("21")
	if err != nil {
		t.Fatalf("Error creando JavaInstaller: %v", err)
	}

	// Verificar que se detectó una distro válida
	validPackageManagers := []string{"apt", "yum", "dnf", "pacman", "apk"}
	valid := false
	for _, pm := range validPackageManagers {
		if installer.packageMgr == pm {
			valid = true
			break
		}
	}

	if !valid {
		t.Errorf("Gestor de paquetes no válido: %s", installer.packageMgr)
	}

	if installer.javaPackage == "" {
		t.Error("Paquete de Java no fue determinado")
	}
}

// TestInstall es más complejo porque requiere permisos de sudo
// Lo marcamos como test de integración que se ejecuta opcionalmente
func TestInstall_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Solo ejecutar si la variable de entorno está configurada
	// Ejemplo: TEST_JAVA_INSTALL=1 go test -run TestInstall_Integration
	// if os.Getenv("TEST_JAVA_INSTALL") != "1" {
	// 	t.Skip("TEST_JAVA_INSTALL no está configurado")
	// }

	// Este test NO se ejecuta por defecto porque:
	// 1. Requiere permisos de sudo
	// 2. Puede modificar el sistema
	// 3. Toma tiempo (descarga e instalación)
	t.Skip("Test de instalación real requiere permisos sudo y modificación del sistema")

	// Descomentar para test manual:
	// installer, err := NewJavaInstaller("21")
	// if err != nil {
	// 	t.Fatalf("Error creando JavaInstaller: %v", err)
	// }
	//
	// err = installer.Install()
	// if err != nil {
	// 	t.Fatalf("Error instalando Java: %v", err)
	// }
	//
	// installed, version, err := installer.CheckInstalled()
	// if !installed {
	// 	t.Error("Java no se instaló correctamente")
	// }
	// t.Logf("Java instalado: %s", version)
}
