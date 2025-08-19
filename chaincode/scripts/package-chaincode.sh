#!/bin/bash

echo "=== LedgerLynx Chaincode Packaging ==="

# Set chaincode details
CHAINCODE_NAME="evidence"
CHAINCODE_VERSION="1.0"
CHAINCODE_PATH="../evidence"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Change to chaincode directory
cd $CHAINCODE_PATH

echo "📦 Packaging chaincode..."
echo "   Name: $CHAINCODE_NAME"
echo "   Version: $CHAINCODE_VERSION"
echo "   Path: $CHAINCODE_PATH"
echo ""

# Build the chaincode to check for errors
echo "🔨 Building chaincode..."
go mod tidy
go build

if [ $? -ne 0 ]; then
    print_error "Chaincode build failed"
    exit 1
fi

print_success "Chaincode built successfully"

# Create package
echo "📦 Creating chaincode package..."
peer lifecycle chaincode package ${CHAINCODE_NAME}.tar.gz --path . --lang golang --label ${CHAINCODE_NAME}_${CHAINCODE_VERSION}

if [ $? -ne 0 ]; then
    print_error "Chaincode packaging failed"
    exit 1
fi

print_success "Chaincode packaged: ${CHAINCODE_NAME}.tar.gz"

# Move package to network directory for deployment
mv ${CHAINCODE_NAME}.tar.gz ../../network/

print_success "Package moved to network directory"
echo ""
echo "🚀 Chaincode ready for deployment!"
echo "   Package: ../../network/${CHAINCODE_NAME}.tar.gz"
echo ""
echo "Next steps:"
echo "1. Install chaincode on all peers"
echo "2. Approve chaincode for each organization"
echo "3. Commit chaincode to channel"
