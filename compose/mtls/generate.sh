#!/bin/bash

# Create directories
mkdir -p {ca,server,client}

# Generate root CA key and certificate
openssl genrsa -out ca/ca.key 4096
openssl req -x509 -new -nodes -key ca/ca.key -sha256 -days 3650 -out ca/ca.crt \
  -subj "/CN=MyTestCA"

# Create extensions file for server cert
cat > server/server.ext << EOF
authorityKeyIdentifier=keyid,issuer
basicConstraints=CA:FALSE
keyUsage = digitalSignature, nonRepudiation, keyEncipherment, dataEncipherment
subjectAltName = @alt_names

[alt_names]
DNS.1 = localhost
DNS.2 = postgres
DNS.3 = mysql
DNS.4 = sqlserver
DNS.5 = 127.0.0.1
DNS.6 = ::1
IP.1 = 127.0.0.1
IP.2 = ::1
EOF

# Generate server key and CSR
openssl genrsa -out server/server.key 2048
openssl req -new -key server/server.key -out server/server.csr \
  -subj "/CN=sqlserver"

# Sign server certificate with CA using extensions
openssl x509 -req -in server/server.csr -CA ca/ca.crt -CAkey ca/ca.key \
  -CAcreateserial -out server/server.crt -days 365 -sha256 \
  -extfile server/server.ext

# Generate client key and CSR
openssl genrsa -out client/client.key 2048
openssl req -new -key client/client.key -out client/client.csr \
  -subj "/CN=testclient"

# Sign client certificate with CA
openssl x509 -req -in client/client.csr -CA ca/ca.crt -CAkey ca/ca.key \
  -CAcreateserial -out client/client.crt -days 365 -sha256

# Set permissions
chmod 600 {ca,server,client}/*.key

# Clean up files
rm server/server.csr server/server.ext client/client.csr
