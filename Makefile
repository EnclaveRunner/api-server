# Makefile for api-server

VERSION ?= v1
API_VERSION ?= api/$(VERSION)
MAIN_GO := main.go
SWAGGER_FILE := docs/swagger.yaml
SWAG_BIN := $(shell go env GOPATH)/bin/swag

set-version:
	@echo "Setting API version to $(API_VERSION) and Swagger BasePath to /$(API_VERSION)"
	# Update APIVersion constant in main.go (robust to whitespace)
	sed -i '' 's|^[[:space:]]*APIVersion[[:space:]]*=[[:space:]]*"[^"]*"|    APIVersion = "$(API_VERSION)"|' $(MAIN_GO)
	# Update @BasePath in main.go (swagger comment)
	sed -i '' 's|^\(// @BasePath\s*\).*|\1/$(API_VERSION)|' $(MAIN_GO)
	if [ -f $(SWAGGER_FILE) ]; then \
		sed -i '' 's|^\(\s*basePath:\s*\).*|\1/$(API_VERSION)|' $(SWAGGER_FILE); \
	fi
	@echo "Regenerating Swagger docs..."
	$(SWAG_BIN) init -g $(MAIN_GO) --output docu

# Spins up containers for development
dev:
	podman-compose -f ./deploy/docker-compose.yml up -d

dev-down:
	podman-compose -f ./deploy/docker-compose.yml down

.PHONY: set-version compose-up compose-down
