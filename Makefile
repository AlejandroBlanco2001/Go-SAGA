KUBECTL := minikube kubectl --

unit-test:
	go test -v ./...

test:
	make unit-test

build-binary-local-order-service:
	echo "Building binary for local order service"
	GOOS=linux GOARCH=amd64 go build -o bin/orders-command cmd/orders-command/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/orders-command-darwin-amd64 cmd/orders-command/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/orders-command-windows-amd64.exe cmd/orders-command/main.go

build-images:
	docker build -t orders-image:latest -f ./docker/orders-command/orders.dockerfile .
	docker build -t inventory-image:latest -f ./docker/inventory-command/inventory.dockerfile .

local-k8s-deploy:
	$(KUBECTL) apply -f k9s/orders/postgres-deployment.yaml
	$(KUBECTL) apply -f k9s/orders/postgres-service.yaml
	$(KUBECTL) apply -f k9s/orders/orders-deployment.yaml
	$(KUBECTL) apply -f k9s/orders/orders-service.yaml

run-local:
	make build-images
	make local-k8s-deploy