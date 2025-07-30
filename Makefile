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
	$(KUBECTL) apply -f k8s/kafka/kafka-stafulset.yml
	$(KUBECTL) apply -f k8s/orders/postgres-deployment.yaml
	$(KUBECTL) apply -f k8s/kafka/kafka-service.yml
	$(KUBECTL) apply -f k8s/orders/postgres-service.yaml
	$(KUBECTL) apply -f k8s/kafka/kafka-ui-deployment.yml
	$(KUBECTL) apply -f k8s/kafka/kafka-topic-creator.yml
	$(KUBECTL) apply -f k8s/orders/orders-service.yaml
	$(KUBECTL) apply -f k8s/orders/orders-deployment.yaml
	$(KUBECTL) apply -f k8s/orders/inventory-service.yaml
	$(KUBECTL) apply -f k8s/orders/inventory-deployment.yaml

local-k8s-delete:
	$(KUBECTL) delete -f k8s/kafka/kafka-ui-deployment.yml --ignore-not-found
	$(KUBECTL) delete -f k8s/kafka/kafka-topic-creator.yml --ignore-not-found
	$(KUBECTL) delete -f k8s/kafka/kafka-stafulset.yml --ignore-not-found
	$(KUBECTL) delete -f k8s/orders/orders-service.yaml --ignore-not-found
	$(KUBECTL) delete -f k8s/orders/orders-deployment.yaml --ignore-not-found
	$(KUBECTL) delete -f k8s/orders/postgres-service.yaml --ignore-not-found
	$(KUBECTL) delete -f k8s/orders/postgres-deployment.yaml --ignore-not-found
	$(KUBECTL) delete -f k8s/orders/inventory-service.yaml --ignore-not-found
	$(KUBECTL) delete -f k8s/orders/inventory-deployment.yaml --ignore-not-found

run-local:
	make build-images
	make local-k8s-deploy