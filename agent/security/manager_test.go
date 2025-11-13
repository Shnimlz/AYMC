package security

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSecurityManager(t *testing.T) {
sm, err := NewSecurityManager("", "")
if err != nil {
t.Fatalf("Error creando SecurityManager: %v", err)
}
if sm == nil {
t.Fatal("SecurityManager es nil")
}
if !sm.HasTLSConfig() {
t.Error("Deberia tener configuracion TLS")
}
}

func TestGenerateToken(t *testing.T) {
sm, err := NewSecurityManager("", "")
if err != nil {
t.Fatalf("Error creando SecurityManager: %v", err)
}
token, err := sm.GenerateToken()
if err != nil {
t.Errorf("Error generando token: %v", err)
}
if len(token) != 64 {
t.Errorf("Token esperado de 64 caracteres, obtenido %d", len(token))
}
}

func TestValidateToken(t *testing.T) {
sm, err := NewSecurityManager("", "")
if err != nil {
t.Fatalf("Error creando SecurityManager: %v", err)
}
validToken := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
if !sm.ValidateToken(validToken) {
t.Error("Token valido fue rechazado")
}
invalidToken := "short"
if sm.ValidateToken(invalidToken) {
t.Error("Token invalido fue aceptado")
}
}

func TestSaveCertificates(t *testing.T) {
sm, err := NewSecurityManager("", "")
if err != nil {
t.Fatalf("Error creando SecurityManager: %v", err)
}
tmpDir := t.TempDir()
certPath := filepath.Join(tmpDir, "cert.pem")
keyPath := filepath.Join(tmpDir, "key.pem")
err = sm.SaveCertificates(certPath, keyPath)
if err != nil {
t.Errorf("Error guardando certificados: %v", err)
}
if _, err := os.Stat(certPath); os.IsNotExist(err) {
t.Error("Archivo de certificado no fue creado")
}
if _, err := os.Stat(keyPath); os.IsNotExist(err) {
t.Error("Archivo de clave no fue creado")
}
}

func TestGetTLSConfig(t *testing.T) {
sm, err := NewSecurityManager("", "")
if err != nil {
t.Fatalf("Error creando SecurityManager: %v", err)
}
tlsConfig := sm.GetTLSConfig()
if tlsConfig == nil {
t.Fatal("TLS Config es nil")
}
if len(tlsConfig.Certificates) == 0 {
t.Error("No hay certificados en TLS Config")
}
}
