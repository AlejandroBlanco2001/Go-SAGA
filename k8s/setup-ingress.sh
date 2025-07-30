#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🚀 Setting up Ingress Hosts for Saga-Go...${NC}"

# Check if running as root or with sudo
if [ "$EUID" -ne 0 ]; then
    echo -e "${YELLOW}⚠️  This script needs sudo privileges to modify /etc/hosts${NC}"
    echo -e "${YELLOW}   Please run: sudo $0${NC}"
    exit 1
fi

# Check if minikube is running
if ! minikube status >/dev/null 2>&1; then
    echo -e "${RED}❌ Minikube is not running. Please start minikube first:${NC}"
    echo -e "${YELLOW}   minikube start${NC}"
    exit 1
fi

# Get minikube IP
echo -e "${BLUE}📡 Getting Minikube IP...${NC}"
MINIKUBE_IP=$(minikube ip)

if [ -z "$MINIKUBE_IP" ]; then
    echo -e "${RED}❌ Could not get Minikube IP. Is minikube running?${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Minikube IP: $MINIKUBE_IP${NC}"

# Check if ingress is deployed
echo -e "${BLUE}🔍 Checking if ingress is deployed...${NC}"
if ! kubectl get ingress saga-go-ingress >/dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Ingress not found. Please deploy your Helm chart first:${NC}"
    echo -e "${YELLOW}   helm install saga-go .${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Ingress found${NC}"

# Hostname to add
HOSTNAME="saga-go.local"

# Check if entry already exists
if grep -q "$HOSTNAME" /etc/hosts; then
    echo -e "${YELLOW}⚠️  Hostname $HOSTNAME already exists in /etc/hosts${NC}"
    echo -e "${BLUE}🔄 Updating existing entry...${NC}"
    
    # Remove existing entry
    sed -i "/$HOSTNAME/d" /etc/hosts
fi

# Add new entry
echo -e "${BLUE}📝 Adding $MINIKUBE_IP $HOSTNAME to /etc/hosts...${NC}"
echo "$MINIKUBE_IP $HOSTNAME" >> /etc/hosts

# Verify the entry was added
if grep -q "$HOSTNAME" /etc/hosts; then
    echo -e "${GREEN}✅ Successfully added $HOSTNAME to /etc/hosts${NC}"
else
    echo -e "${RED}❌ Failed to add entry to /etc/hosts${NC}"
    exit 1
fi

# Test the setup
echo -e "${BLUE}🧪 Testing ingress setup...${NC}"
sleep 2

# Test if the hostname resolves
if ping -c 1 "$HOSTNAME" >/dev/null 2>&1; then
    echo -e "${GREEN}✅ Hostname $HOSTNAME resolves correctly${NC}"
else
    echo -e "${YELLOW}⚠️  Hostname resolution test failed (this might be normal)${NC}"
fi

# Test HTTP connectivity
if curl -s -o /dev/null -w "%{http_code}" "http://$HOSTNAME/kafka-ui" | grep -q "200\|404"; then
    echo -e "${GREEN}✅ HTTP connectivity to ingress is working${NC}"
else
    echo -e "${YELLOW}⚠️  HTTP connectivity test failed${NC}"
fi

echo -e "${GREEN}🎉 Ingress setup complete!${NC}"
echo -e "${BLUE}📋 You can now access your services at:${NC}"
echo -e "${YELLOW}   • Orders API: http://$HOSTNAME/orders${NC}"
echo -e "${YELLOW}   • Inventory API: http://$HOSTNAME/inventory${NC}"
echo -e "${YELLOW}   • Kafka UI: http://$HOSTNAME/kafka-ui${NC}"
echo ""
echo -e "${BLUE}💡 To remove the hostname later, run:${NC}"
echo -e "${YELLOW}   sudo sed -i '/$HOSTNAME/d' /etc/hosts${NC}" 