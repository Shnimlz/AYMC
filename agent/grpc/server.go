package grpc

import (
	"log"
	"net"

	"github.com/aymc/agent/core"
	pb "github.com/aymc/agent/proto"
	"github.com/aymc/agent/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Server es el servidor gRPC del agente
type Server struct {
	agent       *core.Agent
	security    *security.SecurityManager
	grpcServer  *grpc.Server
	listener    net.Listener
}

// NewServer crea un nuevo servidor gRPC
func NewServer(agent *core.Agent, secManager *security.SecurityManager) *Server {
	var opts []grpc.ServerOption

	// Configurar TLS si está disponible
	if secManager.HasTLSConfig() {
		creds := credentials.NewTLS(secManager.GetTLSConfig())
		opts = append(opts, grpc.Creds(creds))
		log.Printf("[INFO] TLS habilitado para gRPC")
	} else {
		log.Printf("[WARN] gRPC ejecutándose sin TLS")
	}

	// Opciones adicionales
	opts = append(opts,
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB
	)

	grpcServer := grpc.NewServer(opts...)

	// Registrar servicio AgentService
	serviceImpl := &agentServiceImpl{agent: agent}
	pb.RegisterAgentServiceServer(grpcServer, serviceImpl)

	// Habilitar reflection para desarrollo
	reflection.Register(grpcServer)

	return &Server{
		agent:      agent,
		security:   secManager,
		grpcServer: grpcServer,
	}
}

// Serve inicia el servidor gRPC
func (s *Server) Serve(listener net.Listener) error {
	s.listener = listener
	log.Printf("[INFO] Servidor gRPC iniciado")
	return s.grpcServer.Serve(listener)
}

// GracefulStop detiene el servidor gracefully
func (s *Server) GracefulStop() {
	log.Printf("[INFO] Deteniendo servidor gRPC...")
	s.grpcServer.GracefulStop()
	log.Printf("[INFO] Servidor gRPC detenido")
}

// Stop detiene el servidor inmediatamente
func (s *Server) Stop() {
	log.Printf("[WARN] Deteniendo servidor gRPC forzadamente...")
	s.grpcServer.Stop()
}

