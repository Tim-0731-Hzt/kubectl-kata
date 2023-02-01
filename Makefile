
export GO111MODULE=on

.PHONY: test
test:
	go test ./pkg/... ./cmd/... -coverprofile cover.out

.PHONY: bin
bin: fmt vet
	go build -o bin/kata github.com/Tim-0731-Hzt/kubectl-kata/cmd/plugin

.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: kubernetes-deps
kubernetes-deps:
	go get k8s.io/client-go@latest
	go get k8s.io/api@latest
	go get k8s.io/apimachinery@latest
	go get k8s.io/cli-runtime@latest

.PHONY: setup
setup:
	make -C setup