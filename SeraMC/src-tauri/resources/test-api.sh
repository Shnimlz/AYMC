#!/bin/bash

# Script de testing para la API de AYMC
# Uso: ./test-api.sh [backend_url]

set -e

# Configuración
BACKEND_URL="${1:-http://localhost:8080}"
API_URL="${BACKEND_URL}/api/v1"
TOKEN_FILE="/tmp/aymc_token.txt"

# Colores para output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Función para imprimir mensajes
print_test() {
    echo -e "${YELLOW}[TEST]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

print_error() {
    echo -e "${RED}[✗]${NC} $1"
}

print_info() {
    echo -e "${NC}[INFO]${NC} $1"
}

# Función para hacer requests
api_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local use_auth=$4
    
    local curl_cmd="curl -s -X $method"
    
    if [ "$use_auth" = "true" ] && [ -f "$TOKEN_FILE" ]; then
        local token=$(cat "$TOKEN_FILE")
        curl_cmd="$curl_cmd -H 'Authorization: Bearer $token'"
    fi
    
    curl_cmd="$curl_cmd -H 'Content-Type: application/json'"
    
    if [ ! -z "$data" ]; then
        curl_cmd="$curl_cmd -d '$data'"
    fi
    
    curl_cmd="$curl_cmd $endpoint"
    
    eval $curl_cmd
}

# Test 1: Health Check
test_health() {
    print_test "Testing health endpoint..."
    response=$(curl -s "${BACKEND_URL}/health")
    
    if echo "$response" | grep -q "healthy"; then
        print_success "Health check passed"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "Health check failed"
        echo "$response"
        return 1
    fi
}

# Test 2: Register User
test_register() {
    print_test "Testing user registration..."
    
    local username="testuser_$(date +%s)"
    local email="test_$(date +%s)@example.com"
    local password="TestPassword123!"
    
    response=$(curl -s -X POST "${API_URL}/auth/register" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"email\":\"$email\",\"password\":\"$password\"}")
    
    if echo "$response" | grep -q "\"id\""; then
        print_success "User registration successful"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        
        # Guardar credenciales para login
        echo "$username" > /tmp/aymc_test_user.txt
        echo "$password" >> /tmp/aymc_test_user.txt
        return 0
    else
        print_error "User registration failed"
        echo "$response"
        return 1
    fi
}

# Test 3: Login
test_login() {
    print_test "Testing login..."
    
    if [ ! -f /tmp/aymc_test_user.txt ]; then
        print_error "No test user found. Run registration first."
        return 1
    fi
    
    local username=$(sed -n '1p' /tmp/aymc_test_user.txt)
    local password=$(sed -n '2p' /tmp/aymc_test_user.txt)
    
    response=$(curl -s -X POST "${API_URL}/auth/login" \
        -H "Content-Type: application/json" \
        -d "{\"username\":\"$username\",\"password\":\"$password\"}")
    
    if echo "$response" | grep -q "\"access_token\""; then
        print_success "Login successful"
        
        # Guardar token
        echo "$response" | jq -r '.access_token' > "$TOKEN_FILE"
        print_info "Token saved to $TOKEN_FILE"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "Login failed"
        echo "$response"
        return 1
    fi
}

# Test 4: Get Profile
test_profile() {
    print_test "Testing get profile..."
    
    if [ ! -f "$TOKEN_FILE" ]; then
        print_error "No token found. Login first."
        return 1
    fi
    
    local token=$(cat "$TOKEN_FILE")
    response=$(curl -s "${API_URL}/auth/me" \
        -H "Authorization: Bearer $token")
    
    if echo "$response" | grep -q "\"username\""; then
        print_success "Get profile successful"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "Get profile failed"
        echo "$response"
        return 1
    fi
}

# Test 5: List Servers
test_list_servers() {
    print_test "Testing list servers..."
    
    if [ ! -f "$TOKEN_FILE" ]; then
        print_error "No token found. Login first."
        return 1
    fi
    
    local token=$(cat "$TOKEN_FILE")
    response=$(curl -s "${API_URL}/servers" \
        -H "Authorization: Bearer $token")
    
    if echo "$response" | grep -q "servers"; then
        print_success "List servers successful"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "List servers failed"
        echo "$response"
        return 1
    fi
}

# Test 6: List Agents
test_list_agents() {
    print_test "Testing list agents..."
    
    if [ ! -f "$TOKEN_FILE" ]; then
        print_error "No token found. Login first."
        return 1
    fi
    
    local token=$(cat "$TOKEN_FILE")
    response=$(curl -s "${API_URL}/agents" \
        -H "Authorization: Bearer $token")
    
    if echo "$response" | grep -q "agents"; then
        print_success "List agents successful"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "List agents failed"
        echo "$response"
        return 1
    fi
}

# Test 7: Search Marketplace
test_search_marketplace() {
    print_test "Testing marketplace search..."
    
    if [ ! -f "$TOKEN_FILE" ]; then
        print_error "No token found. Login first."
        return 1
    fi
    
    local token=$(cat "$TOKEN_FILE")
    response=$(curl -s "${API_URL}/marketplace/search?query=worldedit&limit=5" \
        -H "Authorization: Bearer $token")
    
    if echo "$response" | grep -q "plugins"; then
        print_success "Marketplace search successful"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "Marketplace search failed"
        echo "$response"
        return 1
    fi
}

# Test 8: Protected Endpoint (should work with token)
test_protected_endpoint() {
    print_test "Testing protected endpoint with valid token..."
    
    if [ ! -f "$TOKEN_FILE" ]; then
        print_error "No token found. Login first."
        return 1
    fi
    
    local token=$(cat "$TOKEN_FILE")
    response=$(curl -s "${API_URL}/protected" \
        -H "Authorization: Bearer $token")
    
    if echo "$response" | grep -q "protected endpoint"; then
        print_success "Protected endpoint access successful"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "Protected endpoint access failed"
        echo "$response"
        return 1
    fi
}

# Test 9: Invalid Token (should fail)
test_invalid_token() {
    print_test "Testing protected endpoint with invalid token..."
    
    response=$(curl -s "${API_URL}/servers" \
        -H "Authorization: Bearer invalid_token_here")
    
    if echo "$response" | grep -q "error"; then
        print_success "Invalid token correctly rejected"
        echo "$response" | jq '.' 2>/dev/null || echo "$response"
        return 0
    else
        print_error "Invalid token was not rejected (security issue!)"
        echo "$response"
        return 1
    fi
}

# Main test runner
main() {
    echo "========================================"
    echo "AYMC API Testing Script"
    echo "========================================"
    echo "Backend URL: $BACKEND_URL"
    echo "========================================"
    echo ""
    
    # Check if jq is installed
    if ! command -v jq &> /dev/null; then
        print_info "jq not found. Install it for better JSON output: sudo pacman -S jq"
    fi
    
    local passed=0
    local failed=0
    
    # Run tests
    echo "Running tests..."
    echo ""
    
    if test_health; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_register; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_login; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_profile; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_list_servers; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_list_agents; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_search_marketplace; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_protected_endpoint; then ((passed++)); else ((failed++)); fi
    echo ""
    
    if test_invalid_token; then ((passed++)); else ((failed++)); fi
    echo ""
    
    # Summary
    echo "========================================"
    echo "Test Summary"
    echo "========================================"
    print_success "Passed: $passed"
    print_error "Failed: $failed"
    echo "========================================"
    
    # Cleanup option
    echo ""
    read -p "Clean up test files? (y/n) " -n 1 -r
    echo ""
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -f "$TOKEN_FILE" /tmp/aymc_test_user.txt
        print_info "Test files cleaned up"
    fi
    
    # Exit with error if any test failed
    if [ $failed -gt 0 ]; then
        exit 1
    fi
}

# Run main
main
