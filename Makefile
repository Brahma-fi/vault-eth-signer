.EXPORT_ALL_VARIABLES:

GO    = go
OS	  ="$(shell go env var GOOS | xargs)"
GOBIN =$(PWD)/.bin
path :=$(if $(path), $(path), "./")

version :=v0.0.3

.PHONY: lint
lint: ## Run linters
	$(info $(M) running linters...)
	golangci-lint run --timeout 5m0s ./...

.PHONY: build-linux-release
build-linux-release:  ## - build a static release linux elf(binary)
	@ CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags='-w -s -extldflags "-static"' -a -o "$(GOBIN)/release/linux/vault-eth-signer-$(version)" cmd/plugin/*.go
	@ ls -lah $(GOBIN)/release/linux/vault-eth-signer-$(version)
	@ shasum -a 256 $(GOBIN)/release/linux/vault-eth-signer-$(version)
	@ mkdir -p ./.build/vault/plugins
	@ cp $(GOBIN)/release/linux/vault-eth-signer-$(version) ./.build/vault/plugins/vault-eth-signer-$(version)

.PHONY: build-common
build-common: ## - execute build common tasks clean and mod tidy
	@ $(GO) version
	@ $(GO) clean
	@ $(GO) mod tidy && $(GO) mod download
	@ $(GO) mod verify

.PHONY: test
test: ## - execute go test command
	@ go test -v -cover ./...

build: build-common ## - build a debug binary to the current platform (windows, linux or darwin(mac))
	@ echo cleaning...
	@ rm -f $(GOBIN)/debug/$(OS)/vault-eth-signer
	@ echo building...
	@ $(GO) build -tags dev -o "$(GOBIN)/debug/$(OS)/vault-eth-signer" cmd/plugin/*.go
	@ ls -lah $(GOBIN)/debug/$(OS)/vault-eth-signer
	@ shasum -a 256 $(GOBIN)/debug/$(OS)/vault-eth-signer

.PHONY: scan
scan: ## - execute static code analysis
	@ gosec -fmt=sarif -out=vault-vault-eth-signer.sarif -exclude-dir=test -exclude-dir=bin -severity=medium ./... | 2>&1
	@ echo ""
	@ cat $(path)/vault-vault-eth-signer.sarif

.PHONY: docker-compose-up
docker-compose-up:
	@ docker-compose -f docker-compose.yml up -d

.PHONY: setup-vault
start-vault: docker-compose-up
	@ ./.scripts/vault-setup.sh "$(version)"

.PHONY: docker-compose-down
docker-compose-down:
	@ docker-compose -f docker-compose.yml down

.PHONY: vault-test
vault-test:
	@ ./.scripts/vault-test.sh

test-coverage: ## - execute go test command with coverage
	@ mkdir -p .coverage && mkdir -p .report
	@ go test -json -v -cover -covermode=atomic -coverprofile=.coverage/cover.out ./... > .report/report.out

.PHONY: sonar-scan-local
sonar-scan-local: test-coverage ## - start sonar qube locally with docker (you will need docker installed in your machine)
	@ $(SHELL) .scripts/sonar-start.sh
	@ echo "login with user: admin pwd: 1234"

.PHONY: sonar-stop
sonar-stop: ## - stop sonar qube docker container
	@ docker stop sonarqube
