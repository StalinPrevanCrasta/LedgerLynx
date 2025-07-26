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


# Function to bring up the Fabric network
networkUp() {
    print_message "Starting Hyperledger Fabric network..."
    
    # Check if network artifacts exist
    if [ ! -f "channel-artifacts/genesis.block" ]; then
        print_error "Network artifacts not found. Run './network.sh setup' first"
        exit 1
    fi
    
    # Start the network
    cd docker
    docker-compose -f docker-compose-fabric.yml up -d
    
    if [ $? -ne 0 ]; then
        print_error "Failed to start Fabric network"
        exit 1
    fi
    
    # Wait for containers to be ready
    print_message "Waiting for containers to be ready..."
    sleep 10
    
    # Verify containers are running
    docker-compose -f docker-compose-fabric.yml ps
    
    print_message "Fabric network started successfully!"
    print_message "Network components:"
    print_message "- Orderer: localhost:7050"
    print_message "- Police Dept Peers: localhost:7051, localhost:8051"
    print_message "- Forensics Lab Peers: localhost:9051, localhost:10051"
    print_message "- Court System Peers: localhost:11051, localhost:12051"
    
    cd ..
}

# Function to bring down the Fabric network
networkDown() {
    print_message "Stopping Hyperledger Fabric network..."
    
    cd docker
    docker-compose -f docker-compose-fabric.yml down --volumes --remove-orphans
    
    # Clean up any leftover containers
    docker container prune -f
    docker volume prune -f
    
    print_message "Fabric network stopped and cleaned up"
    cd ..
}

# Function to restart the network
networkRestart() {
    print_message "Restarting Hyperledger Fabric network..."
    networkDown
    sleep 5
    networkUp
}

# Update the main case statement
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
    "up")
        networkUp
        ;;
    "down")
        networkDown
        ;;
    "restart")
        networkRestart
        ;;
    *)
        echo "Usage: $0 {crypto|genesis|channel|setup|up|down|restart}"
        echo "  crypto   - Generate cryptographic material"
        echo "  genesis  - Generate genesis block"
        echo "  channel  - Generate channel configuration"
        echo "  setup    - Complete network setup"
        echo "  up       - Start the Fabric network"
        echo "  down     - Stop the Fabric network"
        echo "  restart  - Restart the Fabric network"
        ;;
esac

