KUBECTL := minikube kubectl --

install-dependencies:
	./setup.sh setup

unit-test:
	go test -v ./...

test:
	make unit-test

dev:
	@echo "Starting application with Skaffold..."
	skaffold dev

lint-helm:
	helm lint ./k8s

dry-run-helm:
	helm install test-release ./k8s --dry-run

delete-helm:
	helm delete test-release

deploy-helm:
	helm install test-release ./k8s
