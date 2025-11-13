package core

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// JavaInstaller maneja la instalación de Java en diferentes sistemas operativos
type JavaInstaller struct {
	version     string
	osType      string
	distro      string
	packageMgr  string
	javaPackage string
}

// NewJavaInstaller crea un nuevo instalador de Java
func NewJavaInstaller(version string) (*JavaInstaller, error) {
	installer := &JavaInstaller{
		version: version,
		osType:  runtime.GOOS,
	}

	// Detectar distribución y gestor de paquetes
	if err := installer.detectSystem(); err != nil {
		return nil, fmt.Errorf("error detectando sistema: %w", err)
	}

	return installer, nil
}

// detectSystem detecta el sistema operativo y el gestor de paquetes
func (ji *JavaInstaller) detectSystem() error {
	switch ji.osType {
	case "linux":
		return ji.detectLinuxDistro()
	case "windows":
		ji.packageMgr = "choco"
		ji.javaPackage = fmt.Sprintf("openjdk%s", ji.version)
		return nil
	case "darwin":
		ji.packageMgr = "brew"
		ji.javaPackage = fmt.Sprintf("openjdk@%s", ji.version)
		return nil
	default:
		return fmt.Errorf("sistema operativo no soportado: %s", ji.osType)
	}
}

// detectLinuxDistro detecta la distribución de Linux y el gestor de paquetes
func (ji *JavaInstaller) detectLinuxDistro() error {
	// Intentar leer /etc/os-release
	cmd := exec.Command("cat", "/etc/os-release")
	output, err := cmd.Output()
	if err != nil {
		// Fallback a lsb_release
		cmd = exec.Command("lsb_release", "-is")
		output, err = cmd.Output()
		if err != nil {
			return fmt.Errorf("no se pudo detectar la distribución de Linux")
		}
	}

	osRelease := strings.ToLower(string(output))

	// Detectar distribución y gestor de paquetes
	switch {
	case strings.Contains(osRelease, "ubuntu"), strings.Contains(osRelease, "debian"):
		ji.distro = "debian"
		ji.packageMgr = "apt"
		ji.javaPackage = fmt.Sprintf("openjdk-%s-jdk", ji.version)

	case strings.Contains(osRelease, "rhel"), strings.Contains(osRelease, "centos"), strings.Contains(osRelease, "fedora"):
		ji.distro = "rhel"
		// Detectar si usa yum o dnf
		if _, err := exec.LookPath("dnf"); err == nil {
			ji.packageMgr = "dnf"
		} else {
			ji.packageMgr = "yum"
		}
		ji.javaPackage = fmt.Sprintf("java-%s-openjdk", ji.version)

	case strings.Contains(osRelease, "arch"):
		ji.distro = "arch"
		ji.packageMgr = "pacman"
		ji.javaPackage = "jdk-openjdk"

	case strings.Contains(osRelease, "alpine"):
		ji.distro = "alpine"
		ji.packageMgr = "apk"
		ji.javaPackage = fmt.Sprintf("openjdk%s", ji.version)

	default:
		return fmt.Errorf("distribución de Linux no reconocida")
	}

	return nil
}

// CheckInstalled verifica si Java ya está instalado
func (ji *JavaInstaller) CheckInstalled() (bool, string, error) {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, "", nil
	}

	version := string(output)
	return true, version, nil
}

// Install instala Java usando el gestor de paquetes apropiado
func (ji *JavaInstaller) Install() error {
	// Verificar si ya está instalado
	installed, version, _ := ji.CheckInstalled()
	if installed {
		return fmt.Errorf("Java ya está instalado: %s", version)
	}

	var cmd *exec.Cmd

	switch ji.packageMgr {
	case "apt":
		// Actualizar repositorios primero
		updateCmd := exec.Command("sudo", "apt-get", "update")
		if err := updateCmd.Run(); err != nil {
			return fmt.Errorf("error actualizando repositorios apt: %w", err)
		}
		cmd = exec.Command("sudo", "apt-get", "install", "-y", ji.javaPackage)

	case "yum", "dnf":
		cmd = exec.Command("sudo", ji.packageMgr, "install", "-y", ji.javaPackage)

	case "pacman":
		cmd = exec.Command("sudo", "pacman", "-S", "--noconfirm", ji.javaPackage)

	case "apk":
		cmd = exec.Command("sudo", "apk", "add", ji.javaPackage)

	case "choco":
		// Verificar si chocolatey está instalado
		if _, err := exec.LookPath("choco"); err != nil {
			return fmt.Errorf("Chocolatey no está instalado. Instálalo desde https://chocolatey.org")
		}
		cmd = exec.Command("choco", "install", ji.javaPackage, "-y")

	case "brew":
		// Verificar si Homebrew está instalado
		if _, err := exec.LookPath("brew"); err != nil {
			return fmt.Errorf("Homebrew no está instalado. Instálalo desde https://brew.sh")
		}
		cmd = exec.Command("brew", "install", ji.javaPackage)

	default:
		return fmt.Errorf("gestor de paquetes no soportado: %s", ji.packageMgr)
	}

	// Ejecutar el comando de instalación
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error instalando Java: %w\nSalida: %s", err, string(output))
	}

	// Verificar que se instaló correctamente
	installed, installedVersion, verifyErr := ji.CheckInstalled()
	if !installed || verifyErr != nil {
		return fmt.Errorf("Java no se instaló correctamente")
	}

	fmt.Printf("✅ Java instalado exitosamente: %s\n", installedVersion)
	return nil
}

// GetInfo devuelve información sobre el instalador
func (ji *JavaInstaller) GetInfo() map[string]string {
	return map[string]string{
		"os":           ji.osType,
		"distro":       ji.distro,
		"package_mgr":  ji.packageMgr,
		"java_package": ji.javaPackage,
		"version":      ji.version,
	}
}

// InstallJavaVersion instala una versión específica de Java
func InstallJavaVersion(version string) error {
	installer, err := NewJavaInstaller(version)
	if err != nil {
		return err
	}

	fmt.Printf("📦 Instalando Java %s...\n", version)
	fmt.Printf("   SO: %s\n", installer.osType)
	if installer.distro != "" {
		fmt.Printf("   Distro: %s\n", installer.distro)
	}
	fmt.Printf("   Gestor: %s\n", installer.packageMgr)
	fmt.Printf("   Paquete: %s\n", installer.javaPackage)

	return installer.Install()
}
