KUBECTL := minikube kubectl --

install-dependencies:
	./setup.sh setup

unit-test:
	go test -v ./...

test:
	make unit-test

build-binary-local-order-service:
	echo "Building binary for local order service"
	GOOS=linux GOARCH=amd64 go build -o bin/orders-command cmd/orders-command/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/orders-command-darwin-amd64 cmd/orders-command/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/orders-command-windows-amd64.exe cmd/orders-command/main.go

# Run with Docker Compose
run-docker:
	@echo "Starting application with Docker Compose..."
	@chmod +x ./tooling/run-app.sh
	./tooling/run-app.sh --detach
	@echo ""
	@echo "🎉 Application started!"
	@echo ""
	@echo "Services available at:"
	@echo "  • Orders API: http://localhost:8080"
	@echo "  • Inventory API: http://localhost:8081"
	@echo "  • Kafka UI: http://localhost:8085"
	@echo ""
	@echo "To stop: make stop-docker"
	@echo "To view logs: make logs-docker"

# Stop Docker Compose
stop-docker:
	@echo "Stopping Docker Compose services..."
	docker compose down
	@echo "✓ Services stopped"

# View Docker Compose logs
logs-docker:
	@echo "Showing Docker Compose logs..."
	docker compose logs -f

# Clean Docker Compose (including volumes)
clean-docker:
	@echo "Cleaning Docker Compose (including volumes)..."
	docker compose down -v
	@echo "✓ Cleanup complete"

# Run with Kubernetes (Skaffold)
dev:
	@echo "Starting application with Skaffold..."
	skaffold dev

# Stop Kubernetes deployment
stop-k8s:
	@echo "Stopping Kubernetes deployment..."
	skaffold delete
	@echo "✓ Kubernetes deployment stopped"

# Show help
help:
	@echo "Available commands:"
	@echo "  install-dependencies  - Check and install required tools"
	@echo "  run-docker           - Start application with Docker Compose"
	@echo "  stop-docker          - Stop Docker Compose services"
	@echo "  logs-docker          - View Docker Compose logs"
	@echo "  clean-docker         - Clean Docker Compose (including volumes)"
	@echo "  dev                  - Start application with Skaffold (Kubernetes)"
	@echo "  stop-k8s             - Stop Kubernetes deployment"
	@echo "  unit-test            - Run unit tests"
	@echo "  test                 - Run all tests"
	@echo "  help                 - Show this help message"