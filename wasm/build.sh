#!/bin/bash

# Build script for WASM fingerprint generator

echo "Building WASM module..."

export GOOS=js
export GOARCH=wasm

go build -o fingerprint.wasm wasm_main.go

if [ $? -eq 0 ]; then
    echo "✓ WASM build successful: fingerprint.wasm"
    
    cp fingerprint.wasm ../client/public/
    echo "✓ Copied fingerprint.wasm to client/public/"

else
    echo "x WASM build failed"
    cd ../wasm
    exit 1
fi
