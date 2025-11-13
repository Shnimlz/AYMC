package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aymc/agent/core"
	"github.com/aymc/agent/grpc"
	"github.com/aymc/agent/security"
)

const (
	Version = "0.1.0"
	AgentName = "AYMC-Agent"
)

var (
	port      = flag.Int("port", 50051, "Puerto gRPC del agente")
	certFile  = flag.String("cert", "", "Ruta al certificado TLS")
	keyFile   = flag.String("key", "", "Ruta a la clave TLS")
	configFile = flag.String("config", "/etc/aymc/agent.json", "Archivo de configuración")
	debug     = flag.Bool("debug", false, "Modo debug")
)

func main() {
	flag.Parse()

	// Banner de inicio
	printBanner()

	// Configurar logger
	if *debug {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetFlags(log.LstdFlags)
	}

	log.Printf("[INFO] Iniciando %s v%s", AgentName, Version)

	// Cargar configuración
	config, err := core.LoadConfig(*configFile)
	if err != nil {
		log.Printf("[WARN] No se pudo cargar configuración desde %s: %v", *configFile, err)
		log.Printf("[INFO] Usando configuración por defecto")
		config = core.DefaultConfig()
	}

	// Inicializar contexto con cancelación
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Inicializar módulo de seguridad
	secManager, err := security.NewSecurityManager(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("[ERROR] Fallo al inicializar seguridad: %v", err)
	}

	// Inicializar core del agente
	agent, err := core.NewAgent(ctx, config)
	if err != nil {
		log.Fatalf("[ERROR] Fallo al inicializar agente: %v", err)
	}

	// Iniciar servidor gRPC
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("[ERROR] Fallo al escuchar en puerto %d: %v", *port, err)
	}

	grpcServer := grpc.NewServer(agent, secManager)
	
	log.Printf("[INFO] Servidor gRPC escuchando en puerto %d", *port)
	
	// Goroutine para servidor gRPC
	go func() {
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("[ERROR] Error en servidor gRPC: %v", err)
		}
	}()

	// Iniciar monitoreo de sistema
	go agent.StartMonitoring(ctx, 5*time.Second)

	// Manejar señales de sistema
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("[INFO] Señal recibida: %v. Cerrando gracefully...", sig)

	// Detener servidor gRPC
	grpcServer.GracefulStop()

	// Detener agente
	agent.Shutdown()

	log.Printf("[INFO] Agente detenido correctamente")
}

func printBanner() {
	banner := `
╔═══════════════════════════════════════════════════╗
║                                                   ║
║     █████╗ ██╗   ██╗███╗   ███╗ ██████╗         ║
║    ██╔══██╗╚██╗ ██╔╝████╗ ████║██╔════╝         ║
║    ███████║ ╚████╔╝ ██╔████╔██║██║              ║
║    ██╔══██║  ╚██╔╝  ██║╚██╔╝██║██║              ║
║    ██║  ██║   ██║   ██║ ╚═╝ ██║╚██████╗         ║
║    ╚═╝  ╚═╝   ╚═╝   ╚═╝     ╚═╝ ╚═════╝         ║
║                                                   ║
║    Advanced Minecraft Control Agent v%s        ║
║    Sistema de control remoto para servidores MC  ║
║                                                   ║
╚═══════════════════════════════════════════════════╝
`
	fmt.Printf(banner, Version)
	fmt.Println()
}
