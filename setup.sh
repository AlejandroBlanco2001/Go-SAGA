#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    local missing_tools=()
    
    if ! command_exists docker; then
        missing_tools+=("Docker")
    else
        print_success "Docker is installed"
    fi
    
    if ! command_exists docker-compose && ! docker compose version >/dev/null 2>&1; then
        missing_tools+=("Docker Compose")
    else
        print_success "Docker Compose is installed"
    fi

    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        echo "Please install the missing tools:"
        for tool in "${missing_tools[@]}"; do
            case $tool in
                "Docker")
                    echo "  Docker: https://docs.docker.com/get-docker/"
                    ;;
                "Docker Compose")
                    echo "  Docker Compose: https://docs.docker.com/compose/install/"
                    ;;
            esac
        done
        exit 1
    fi
    
    # Optional tools
    if ! command_exists minikube; then
        print_warning "Minikube not found. Install for k8s development: https://minikube.sigs.k8s.io/docs/start/"
    else
        print_success "Minikube is installed"
    fi
    
    if ! command_exists kubectl; then
        print_warning "kubectl not found. Install for k8s development: https://kubernetes.io/docs/tasks/tools/"
    else
        print_success "kubectl is installed"
    fi
    
    if ! command_exists skaffold; then
        print_warning "Skaffold not found. Install for k8s development: https://skaffold.dev/docs/install/"
    else
        print_success "Skaffold is installed"
    fi
}

# Function to show help
show_help() {
    echo "SAGA Pattern Go Application Setup Script"
    echo ""
    echo "Usage: $0 [COMMAND]"
    echo ""
    echo "Commands:"
    echo "  setup         - Complete setup (check deps, install, start)"
    echo "  check-deps    - Check if required tools are installed"
    echo "  help          - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 setup           # Complete setup and start"
    echo ""
}

# Main script logic
case "${1:-help}" in
    "setup")
        check_prerequisites
        ;;
    "check-deps")
        check_prerequisites
        ;;
    "help"|*)
        show_help
        ;;
esac 