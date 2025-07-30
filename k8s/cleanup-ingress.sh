#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 Cleaning up Ingress Hosts for Saga-Go...${NC}"

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}⚠️  This script needs sudo privileges to modify /etc/hosts${NC}"
    echo -e "${YELLOW}   Please run: sudo $0${NC}"
    exit 1
fi

# Hostname to remove
HOSTNAME="saga-go.local"

# Check if entry exists
if grep -q "$HOSTNAME" /etc/hosts; then
    echo -e "${BLUE}🗑️  Removing $HOSTNAME from /etc/hosts...${NC}"
    
    # Remove the entry
    sed -i "/$HOSTNAME/d" /etc/hosts
    
    # Verify removal
    if ! grep -q "$HOSTNAME" /etc/hosts; then
        echo -e "${GREEN}✅ Successfully removed $HOSTNAME from /etc/hosts${NC}"
    else
        echo -e "${RED}❌ Failed to remove entry from /etc/hosts${NC}"
        exit 1
    fi
else
    echo -e "${YELLOW}ℹ️  Hostname $HOSTNAME not found in /etc/hosts${NC}"
fi

echo -e "${GREEN}🎉 Cleanup complete!${NC}"
echo -e "${BLUE}💡 To set up ingress again, run:${NC}"
echo -e "${YELLOW}   sudo ./setup-ingress.sh${NC}" 