#!/bin/bash
# Generate TLS certificates for gRPC services
# Supports both simple TLS (server-only) and mTLS (mutual authentication)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
CERTS_DIR="$PROJECT_ROOT/certs"

# Configuration
DAYS_VALID=3650  # 10 years (for private network)
COUNTRY="US"
STATE="California"
CITY="San Francisco"
ORG="Arduino Trader"
OU="Microservices"

echo "==================================================================="
echo "Arduino Trader - TLS Certificate Generator"
echo "==================================================================="
echo ""

# Parse arguments
MTLS=false
if [ "$1" = "--mtls" ]; then
    MTLS=true
    echo "Mode: mTLS (Mutual TLS) - Server + Client authentication"
else
    echo "Mode: TLS (Server-only authentication)"
    echo "  Use --mtls flag for mutual authentication"
fi
echo ""

# Create certificates directory
mkdir -p "$CERTS_DIR"
cd "$CERTS_DIR"

echo "Certificates will be stored in: $CERTS_DIR"
echo ""

# ============================================================================
# Step 1: Generate Certificate Authority (CA)
# ============================================================================
echo "[1/5] Generating Certificate Authority (CA)..."

if [ ! -f ca-key.pem ] || [ ! -f ca-cert.pem ]; then
    # Generate CA private key
    openssl genrsa -out ca-key.pem 4096

    # Generate CA certificate (self-signed)
    openssl req -new -x509 -days $DAYS_VALID -key ca-key.pem -out ca-cert.pem \
        -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$OU/CN=Arduino Trader CA"

    echo "  ✓ CA certificate created (ca-cert.pem)"
else
    echo "  ✓ CA certificate already exists (ca-cert.pem)"
fi
echo ""

# ============================================================================
# Step 2: Generate Server Certificate
# ============================================================================
echo "[2/5] Generating Server certificate..."

if [ ! -f server-key.pem ] || [ ! -f server-cert.pem ]; then
    # Generate server private key
    openssl genrsa -out server-key.pem 4096

    # Generate server certificate signing request (CSR)
    openssl req -new -key server-key.pem -out server.csr \
        -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$OU/CN=arduino-trader-server"

    # Create extensions file for Subject Alternative Names (SAN)
    cat > server-ext.cnf <<EOF
subjectAltName=DNS:localhost,DNS:*.local,IP:127.0.0.1,IP:192.168.1.100,IP:192.168.1.101
extendedKeyUsage=serverAuth
EOF

    # Sign server certificate with CA
    openssl x509 -req -days $DAYS_VALID -in server.csr \
        -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial \
        -out server-cert.pem -extfile server-ext.cnf

    # Cleanup
    rm server.csr server-ext.cnf

    echo "  ✓ Server certificate created (server-cert.pem)"
else
    echo "  ✓ Server certificate already exists (server-cert.pem)"
fi
echo ""

# ============================================================================
# Step 3: Generate Client Certificates (for mTLS)
# ============================================================================
if [ "$MTLS" = true ]; then
    echo "[3/5] Generating Client certificates (mTLS)..."

    # Device 1 client cert
    if [ ! -f device1-client-key.pem ] || [ ! -f device1-client-cert.pem ]; then
        openssl genrsa -out device1-client-key.pem 4096

        openssl req -new -key device1-client-key.pem -out device1-client.csr \
            -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$OU/CN=device1-client"

        cat > device1-client-ext.cnf <<EOF
extendedKeyUsage=clientAuth
EOF

        openssl x509 -req -days $DAYS_VALID -in device1-client.csr \
            -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial \
            -out device1-client-cert.pem -extfile device1-client-ext.cnf

        rm device1-client.csr device1-client-ext.cnf
        echo "  ✓ Device1 client certificate created"
    else
        echo "  ✓ Device1 client certificate already exists"
    fi

    # Device 2 client cert
    if [ ! -f device2-client-key.pem ] || [ ! -f device2-client-cert.pem ]; then
        openssl genrsa -out device2-client-key.pem 4096

        openssl req -new -key device2-client-key.pem -out device2-client.csr \
            -subj "/C=$COUNTRY/ST=$STATE/L=$CITY/O=$ORG/OU=$OU/CN=device2-client"

        cat > device2-client-ext.cnf <<EOF
extendedKeyUsage=clientAuth
EOF

        openssl x509 -req -days $DAYS_VALID -in device2-client.csr \
            -CA ca-cert.pem -CAkey ca-key.pem -CAcreateserial \
            -out device2-client-cert.pem -extfile device2-client-ext.cnf

        rm device2-client.csr device2-client-ext.cnf
        echo "  ✓ Device2 client certificate created"
    else
        echo "  ✓ Device2 client certificate already exists"
    fi
else
    echo "[3/5] Skipping client certificates (TLS mode)"
fi
echo ""

# ============================================================================
# Step 4: Set Proper Permissions
# ============================================================================
echo "[4/5] Setting file permissions..."

# Private keys should be readable only by owner
chmod 600 *-key.pem
chmod 644 *-cert.pem

echo "  ✓ Private keys: 600 (owner read/write only)"
echo "  ✓ Certificates: 644 (readable by all)"
echo ""

# ============================================================================
# Step 5: Verify Certificates
# ============================================================================
echo "[5/5] Verifying certificates..."

# Verify server certificate
openssl verify -CAfile ca-cert.pem server-cert.pem > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "  ✓ Server certificate valid"
else
    echo "  ✗ Server certificate verification failed"
    exit 1
fi

# Verify client certificates if mTLS
if [ "$MTLS" = true ]; then
    openssl verify -CAfile ca-cert.pem device1-client-cert.pem > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "  ✓ Device1 client certificate valid"
    fi

    openssl verify -CAfile ca-cert.pem device2-client-cert.pem > /dev/null 2>&1
    if [ $? -eq 0 ]; then
        echo "  ✓ Device2 client certificate valid"
    fi
fi
echo ""

# ============================================================================
# Summary
# ============================================================================
echo "==================================================================="
echo "Certificate Generation Complete!"
echo "==================================================================="
echo ""
echo "Generated files in $CERTS_DIR:"
echo ""
echo "Certificate Authority:"
echo "  - ca-cert.pem         (CA certificate - distribute to all devices)"
echo "  - ca-key.pem          (CA private key - KEEP SECURE)"
echo ""
echo "Server Certificates:"
echo "  - server-cert.pem     (Server certificate)"
echo "  - server-key.pem      (Server private key - KEEP SECURE)"
echo ""

if [ "$MTLS" = true ]; then
    echo "Client Certificates (mTLS):"
    echo "  - device1-client-cert.pem  (Device1 client certificate)"
    echo "  - device1-client-key.pem   (Device1 client key - KEEP SECURE)"
    echo "  - device2-client-cert.pem  (Device2 client certificate)"
    echo "  - device2-client-key.pem   (Device2 client key - KEEP SECURE)"
    echo ""
fi

echo "Next Steps:"
echo ""
echo "1. Copy certificates to each device:"
echo "   Device 1 (server):"
echo "     - ca-cert.pem"
echo "     - server-cert.pem"
echo "     - server-key.pem"
if [ "$MTLS" = true ]; then
    echo "     - device1-client-cert.pem (for client calls)"
    echo "     - device1-client-key.pem"
fi
echo ""
echo "   Device 2 (client):"
echo "     - ca-cert.pem"
if [ "$MTLS" = true ]; then
    echo "     - device2-client-cert.pem"
    echo "     - device2-client-key.pem"
fi
echo ""
echo "2. Update services.yaml to enable TLS:"
echo "   tls:"
echo "     enabled: true"
if [ "$MTLS" = true ]; then
    echo "     mutual: true"
fi
echo "     ca_cert: certs/ca-cert.pem"
echo "     server_cert: certs/server-cert.pem"
echo "     server_key: certs/server-key.pem"
if [ "$MTLS" = true ]; then
    echo "     client_cert: certs/device1-client-cert.pem  # or device2"
    echo "     client_key: certs/device1-client-key.pem    # or device2"
fi
echo ""
echo "3. Restart all services"
echo ""
echo "Security Notes:"
echo "  - These certificates are self-signed (suitable for private networks)"
echo "  - For production, consider proper CA-signed certificates"
echo "  - Keep all *-key.pem files secure and never commit to git"
echo "  - Valid for: $DAYS_VALID days (~10 years)"
echo ""

# Create .gitignore for certs directory
cat > .gitignore <<EOF
# Ignore all certificate files (security)
*.pem
*.csr
*.srl
*.cnf

# Keep only the generation script
!README.md
EOF

echo "✓ Created .gitignore to prevent committing certificates"
echo ""
echo "==================================================================="
