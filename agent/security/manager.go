package security

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"
)

// SecurityManager maneja la seguridad del agente
type SecurityManager struct {
	certFile   string
	keyFile    string
	tlsConfig  *tls.Config
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

// NewSecurityManager crea un nuevo gestor de seguridad
func NewSecurityManager(certFile, keyFile string) (*SecurityManager, error) {
	sm := &SecurityManager{
		certFile: certFile,
		keyFile:  keyFile,
	}

	// Si se proporcionan certificados, cargarlos
	if certFile != "" && keyFile != "" {
		if err := sm.loadTLSConfig(); err != nil {
			return nil, fmt.Errorf("error cargando certificados TLS: %w", err)
		}
		log.Printf("[INFO] Certificados TLS cargados desde: %s, %s", certFile, keyFile)
	} else {
		// Generar certificados autofirmados
		log.Printf("[INFO] Generando certificados autofirmados...")
		if err := sm.generateSelfSignedCert(); err != nil {
			return nil, fmt.Errorf("error generando certificados: %w", err)
		}
	}

	return sm, nil
}

// loadTLSConfig carga la configuración TLS desde archivos
func (sm *SecurityManager) loadTLSConfig() error {
	cert, err := tls.LoadX509KeyPair(sm.certFile, sm.keyFile)
	if err != nil {
		return err
	}

	sm.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
	}

	return nil
}

// generateSelfSignedCert genera un certificado autofirmado
func (sm *SecurityManager) generateSelfSignedCert() error {
	// Generar clave privada RSA
	privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return fmt.Errorf("error generando clave privada: %w", err)
	}
	sm.privateKey = privateKey
	sm.publicKey = &privateKey.PublicKey

	// Crear plantilla de certificado
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return fmt.Errorf("error generando número serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"AYMC Agent"},
			CommonName:   "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour), // 1 año
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
	}

	// Crear certificado autofirmado
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, sm.publicKey, sm.privateKey)
	if err != nil {
		return fmt.Errorf("error creando certificado: %w", err)
	}

	// Guardar certificado en memoria (no en disco por seguridad)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(sm.privateKey),
	})

	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return fmt.Errorf("error cargando par de claves: %w", err)
	}

	sm.tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS13,
		CipherSuites: []uint16{
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_AES_128_GCM_SHA256,
		},
	}

	log.Printf("[INFO] Certificado autofirmado generado correctamente")
	return nil
}

// HasTLSConfig verifica si hay configuración TLS disponible
func (sm *SecurityManager) HasTLSConfig() bool {
	return sm.tlsConfig != nil
}

// GetTLSConfig retorna la configuración TLS
func (sm *SecurityManager) GetTLSConfig() *tls.Config {
	return sm.tlsConfig
}

// SaveCertificates guarda los certificados en disco
func (sm *SecurityManager) SaveCertificates(certPath, keyPath string) error {
	if sm.privateKey == nil {
		return fmt.Errorf("no hay clave privada para guardar")
	}

	// Crear directorios si no existen
	if err := os.MkdirAll(certPath[:len(certPath)-len("/cert.pem")], 0700); err != nil {
		return err
	}

	// Guardar certificado
	certFile, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certFile.Close()

	// Obtener certificado del TLS config
	if len(sm.tlsConfig.Certificates) == 0 {
		return fmt.Errorf("no hay certificados disponibles")
	}

	certPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "CERTIFICATE",
		Bytes: sm.tlsConfig.Certificates[0].Certificate[0],
	})

	if _, err := certFile.Write(certPEM); err != nil {
		return err
	}

	// Guardar clave privada
	keyFile, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyFile.Close()

	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(sm.privateKey),
	})

	if _, err := keyFile.Write(keyPEM); err != nil {
		return err
	}

	log.Printf("[INFO] Certificados guardados en: %s, %s", certPath, keyPath)
	return nil
}

// GenerateToken genera un token de autenticación
func (sm *SecurityManager) GenerateToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", tokenBytes), nil
}

// ValidateToken valida un token de autenticación
func (sm *SecurityManager) ValidateToken(token string) bool {
	// TODO: Implementar validación real de tokens
	return len(token) == 64 // Validación básica
}
