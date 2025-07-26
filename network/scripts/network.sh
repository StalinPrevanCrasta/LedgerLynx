#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored output
print_message() {
    echo -e "${GREEN}[LedgerLynx]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to generate crypto material
generateCrypto() {
    print_message "Generating cryptographic material..."
    
    # Check if cryptogen exists
    which cryptogen
    if [ "$?" -ne 0 ]; then
        print_error "cryptogen tool not found. Make sure Hyperledger Fabric binaries are installed"
        exit 1
    fi
    
    # Remove existing crypto material
    rm -rf organizations/
    
    # Generate crypto material
    cryptogen generate --config=./crypto-config.yaml --output="organizations"
    
    if [ "$?" -ne 0 ]; then
        print_error "Failed to generate crypto material"
        exit 1
    fi
    
    print_message "Crypto material generated successfully"
}

# Function to generate genesis block
generateGenesisBlock() {
    print_message "Generating genesis block..."
    
    # Check if configtxgen exists
    which configtxgen
    if [ "$?" -ne 0 ]; then
        print_error "configtxgen tool not found. Make sure Hyperledger Fabric binaries are installed"
        exit 1
    fi
    
    # Set fabric config path
    export FABRIC_CFG_PATH=$PWD
    
    # Create channel-artifacts directory
    mkdir -p channel-artifacts
    
    # Generate genesis block
    configtxgen -profile ThreeOrgsOrdererGenesis -channelID system-channel -outputBlock ./channel-artifacts/genesis.block
    
    if [ "$?" -ne 0 ]; then
        print_error "Failed to generate genesis block"
        exit 1
    fi
    
    print_message "Genesis block generated successfully"
}

# Function to generate channel configuration
generateChannelConfig() {
    print_message "Generating channel configuration..."
    
    export FABRIC_CFG_PATH=$PWD
    
    # Generate channel configuration transaction
    configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx ./channel-artifacts/evidencechannel.tx -channelID evidencechannel
    
    if [ "$?" -ne 0 ]; then
        print_error "Failed to generate channel configuration"
        exit 1
    fi
    
    # Generate anchor peer transactions for each organization
    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/PoliceDeptMSPanchors.tx -channelID evidencechannel -asOrg PoliceDeptMSP
    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/ForensicsLabMSPanchors.tx -channelID evidencechannel -asOrg ForensicsLabMSP
    configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/CourtSystemMSPanchors.tx -channelID evidencechannel -asOrg CourtSystemMSP
    
    print_message "Channel configuration generated successfully"
}

# Function to setup the complete network
setupNetwork() {
    print_message "Setting up LedgerLynx Hyperledger Fabric network..."
    
    # Generate all necessary artifacts
    generateCrypto
    generateGenesisBlock
    generateChannelConfig
    
    print_message "Network setup completed successfully!"
    print_message "Next steps:"
    print_message "1. Start the network with Docker Compose"
    print_message "2. Create and join the evidence channel"
    print_message "3. Deploy evidence management chaincode"
}

# Main execution
case $1 in
    "crypto")
        generateCrypto
        ;;
    "genesis")
        generateGenesisBlock
        ;;
    "channel")
        generateChannelConfig
        ;;
    "setup")
        setupNetwork
        ;;
    *)
        echo "Usage: $0 {crypto|genesis|channel|setup}"
        echo "  crypto  - Generate cryptographic material"
        echo "  genesis - Generate genesis block"
        echo "  channel - Generate channel configuration"
        echo "  setup   - Complete network setup"
        ;;
esac
