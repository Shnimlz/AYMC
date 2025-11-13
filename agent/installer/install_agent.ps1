# AYMC Agent Installer - PowerShell Script para Windows
# Este script instala y configura el agente AYMC en sistemas Windows

param(
    [string]$InstallPath = "C:\Program Files\AYMC",
    [string]$ConfigPath = "C:\ProgramData\AYMC",
    [switch]$NoService = $false,
    [switch]$NoJava = $false
)

$ErrorActionPreference = "Stop"

# Variables
$AGENT_VERSION = "0.1.0"
$SERVICE_NAME = "AYMCAgent"
$DOWNLOAD_URL = "https://github.com/aymc/agent/releases/download/v$AGENT_VERSION/aymc-agent-windows-amd64.exe"

# Funciones auxiliares
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Check-Prerequisites {
    Write-Info "Verificando prerequisitos..."
    
    if (-not (Test-Administrator)) {
        Write-Error-Custom "Este script debe ejecutarse como Administrador"
        exit 1
    }
    
    Write-Info "Ejecutando como Administrador ✓"
}

function Install-Java {
    if ($NoJava) {
        Write-Info "Instalación de Java omitida (flag -NoJava)"
        return
    }
    
    Write-Info "Verificando instalación de Java..."
    
    try {
        $javaVersion = java -version 2>&1 | Select-String "version"
        if ($javaVersion) {
            Write-Info "Java ya está instalado: $javaVersion"
            return
        }
    } catch {
        Write-Info "Java no encontrado"
    }
    
    Write-Warn "Java no está instalado"
    Write-Info "Por favor, descargue e instale Java desde:"
    Write-Info "https://adoptium.net/temurin/releases/"
    Write-Info ""
    
    $response = Read-Host "¿Desea continuar sin Java? (s/N)"
    if ($response -ne "s" -and $response -ne "S") {
        exit 1
    }
}

function Create-Directories {
    Write-Info "Creando directorios..."
    
    $directories = @(
        $InstallPath,
        "$ConfigPath",
        "$InstallPath\servers",
        "$ConfigPath\logs"
    )
    
    foreach ($dir in $directories) {
        if (-not (Test-Path $dir)) {
            New-Item -ItemType Directory -Path $dir -Force | Out-Null
            Write-Info "Creado: $dir"
        }
    }
}

function Download-Agent {
    Write-Info "Descargando agente AYMC v$AGENT_VERSION..."
    
    $agentPath = Join-Path $InstallPath "aymc-agent.exe"
    
    try {
        # Por ahora placeholder
        # Invoke-WebRequest -Uri $DOWNLOAD_URL -OutFile $agentPath
        Write-Warn "Descarga omitida - usando binario local para desarrollo"
    } catch {
        Write-Error-Custom "Error descargando agente: $_"
        exit 1
    }
}

function Create-Config {
    Write-Info "Creando configuración por defecto..."
    
    $javaPath = (Get-Command java -ErrorAction SilentlyContinue).Source
    if (-not $javaPath) {
        $javaPath = "java"
    }
    
    $agentId = "agent-" + (Get-Date -Format "yyyyMMddHHmmss")
    
    $config = @{
        agent_id = $agentId
        backend_url = "localhost:50050"
        port = 50051
        log_level = "info"
        max_servers = 10
        java_path = $javaPath
        work_dir = "$InstallPath\servers"
        enable_metrics = $true
        metrics_interval = "5s"
        custom_env = @{}
    }
    
    $configPath = Join-Path $ConfigPath "agent.json"
    $config | ConvertTo-Json -Depth 10 | Set-Content -Path $configPath -Encoding UTF8
    
    Write-Info "Configuración creada: $configPath"
}

function Install-Service {
    if ($NoService) {
        Write-Info "Instalación de servicio omitida (flag -NoService)"
        return
    }
    
    Write-Info "Instalando servicio de Windows..."
    
    $agentPath = Join-Path $InstallPath "aymc-agent.exe"
    $configFile = Join-Path $ConfigPath "agent.json"
    
    # Usar NSSM o crear servicio directamente
    try {
        # Verificar si el servicio ya existe
        $existingService = Get-Service -Name $SERVICE_NAME -ErrorAction SilentlyContinue
        if ($existingService) {
            Write-Info "Servicio existente encontrado, eliminando..."
            Stop-Service -Name $SERVICE_NAME -Force -ErrorAction SilentlyContinue
            sc.exe delete $SERVICE_NAME | Out-Null
            Start-Sleep -Seconds 2
        }
        
        # Crear servicio
        $binPath = "`"$agentPath`" --config=`"$configFile`""
        sc.exe create $SERVICE_NAME binPath= $binPath start= auto DisplayName= "AYMC Agent" | Out-Null
        sc.exe description $SERVICE_NAME "Advanced Minecraft Control Agent - Gestión remota de servidores Minecraft" | Out-Null
        
        Write-Info "Servicio instalado correctamente"
    } catch {
        Write-Error-Custom "Error instalando servicio: $_"
        Write-Warn "Puede ejecutar el agente manualmente desde: $agentPath"
    }
}

function Configure-Firewall {
    Write-Info "Configurando firewall..."
    
    try {
        $ruleName = "AYMC Agent gRPC"
        
        # Verificar si la regla existe
        $existingRule = Get-NetFirewallRule -DisplayName $ruleName -ErrorAction SilentlyContinue
        if ($existingRule) {
            Write-Info "Regla de firewall ya existe"
            return
        }
        
        # Crear regla
        New-NetFirewallRule -DisplayName $ruleName `
                           -Direction Inbound `
                           -Protocol TCP `
                           -LocalPort 50051 `
                           -Action Allow `
                           -Description "Puerto gRPC para AYMC Agent" | Out-Null
        
        Write-Info "Regla de firewall creada"
    } catch {
        Write-Warn "No se pudo configurar el firewall automáticamente"
        Write-Info "Configure manualmente el puerto 50051/TCP"
    }
}

function Print-Summary {
    Write-Host ""
    Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║                                                            ║" -ForegroundColor Cyan
    Write-Host "║  ✓ AYMC Agent instalado correctamente                     ║" -ForegroundColor Cyan
    Write-Host "║                                                            ║" -ForegroundColor Cyan
    Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Directorios:"
    Write-Host "  - Instalación: $InstallPath"
    Write-Host "  - Configuración: $ConfigPath"
    Write-Host ""
    
    if (-not $NoService) {
        Write-Host "Comandos útiles:"
        Write-Host "  - Iniciar servicio:  Start-Service $SERVICE_NAME"
        Write-Host "  - Detener servicio:  Stop-Service $SERVICE_NAME"
        Write-Host "  - Estado:            Get-Service $SERVICE_NAME"
        Write-Host "  - Ver logs:          Get-Content '$ConfigPath\logs\agent.log' -Wait"
    } else {
        Write-Host "Iniciar manualmente:"
        Write-Host "  & '$InstallPath\aymc-agent.exe' --config='$ConfigPath\agent.json'"
    }
    Write-Host ""
}

# Main
function Main {
    Write-Host "╔════════════════════════════════════════════════════════════╗" -ForegroundColor Cyan
    Write-Host "║     AYMC Agent Installer v$AGENT_VERSION                          ║" -ForegroundColor Cyan
    Write-Host "╚════════════════════════════════════════════════════════════╝" -ForegroundColor Cyan
    Write-Host ""
    
    Check-Prerequisites
    Install-Java
    Create-Directories
    Download-Agent
    Create-Config
    Install-Service
    Configure-Firewall
    Print-Summary
    
    Write-Info "Instalación completada"
}

# Ejecutar
Main
