#!/bin/sh

# Check if both key and filename are provided
if [ $# -lt 2 ]; then
    echo "Usage: $0 <private_key> <filename>"
    exit 1
fi

inkey="$1"
filename="$2"

# Check if file exists
if [ ! -f "$filename" ]; then
    echo "File not found: $filename"
    exit 1
fi

# Sign the file
openssl pkeyutl -sign -inkey "$inkey" -out "${filename}.sig" -rawin -in "$filename"

# Base64 encode the file contents
contents_base64=$(openssl base64 -A < "$filename")

# Base64 encode the signature
signature_base64=$(openssl base64 -A < "${filename}.sig")

# Create JSON structure
json_content="{\"license\":\"$contents_base64\",\"signature\":\"$signature_base64\"}"

# Write JSON to ee_license file
echo "$json_content" > ee_license.json

# Base64 encode the ee_license file and output to stdout
openssl base64 -A < ee_license.json

# Clean up temporary signature file
rm "${filename}.sig"
rm ee_license.json
