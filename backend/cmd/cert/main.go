package main

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"

	_ "github.com/lib/pq"
)

// readAndEncodeBase64 reads the contents of a file and returns the base64 encoded string
func readAndEncodeBase64(filePath string) (string, error) {
	fileContents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}
	encoded := base64.StdEncoding.EncodeToString(fileContents)
	return encoded, nil
}

func main() {
	// Paths to the certificate files
	rootCertPath := "certs/root.crt"
	clientCertPath := "certs/client.crt"
	clientKeyPath := "certs/client.key"

	// Base64 encode the certificate files
	rootCertBase64, err := readAndEncodeBase64(rootCertPath)
	if err != nil {
		log.Fatalf("failed to base64 encode root certificate: %v", err)
	}
	_ = rootCertBase64

	clientCertBase64, err := readAndEncodeBase64(clientCertPath)
	if err != nil {
		log.Fatalf("failed to base64 encode client certificate: %v", err)
	}
	_ = clientCertBase64

	clientKeyBase64, err := readAndEncodeBase64(clientKeyPath)
	if err != nil {
		log.Fatalf("failed to base64 encode client key: %v", err)
	}
	_ = clientKeyBase64

	// Connection string with the custom sslmode and sslconfig parameters
	connStr := fmt.Sprintf("postgres://myuser:mypassword@localhost:5555/mydb?sslmode=require&sslrootcert=%s&sslcert=%s&sslkey=%s", rootCertBase64, clientCertPath, clientKeyPath)

	// Connect to the PostgreSQL database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatalf("failed to open database connection: %v", err)
	}
	defer db.Close()

	// Verify the connection
	err = db.Ping()
	if err != nil {
		log.Fatalf("failed to verify connection: %v", err)
	}

	fmt.Println("Successfully connected to the database!")
}
