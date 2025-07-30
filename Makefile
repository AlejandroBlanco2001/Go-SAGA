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

run-local:
	skaffold dev