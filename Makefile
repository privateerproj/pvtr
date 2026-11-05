# Ref: https://www.digitalocean.com/community/tutorials/using-ldflags-to-set-version-information-for-go-applications

BUILD_FLAGS=-X 'main.GitCommitHash=`git rev-parse --short HEAD`' -X 'main.BuiltAt=`date +%FT%T%z`' -X 'main.Version=`git describe --tags`'
BUILD_WIN_AMD64=@env GOOS=windows GOARCH=amd64 go build -o pvtr-windows-amd64.exe
BUILD_WIN_ARM64=@env GOOS=windows GOARCH=arm64 go build -o pvtr-windows-arm64.exe
BUILD_LINUX_AMD64=@env GOOS=linux GOARCH=amd64 go build -o pvtr-linux-amd64
BUILD_LINUX_ARM64=@env GOOS=linux GOARCH=arm64 go build -o pvtr-linux-arm64
BUILD_MAC_AMD64=@env GOOS=darwin GOARCH=amd64 go build -o pvtr-darwin-amd64
BUILD_MAC_ARM64=@env GOOS=darwin GOARCH=arm64 go build -o pvtr-darwin-arm64

# Detect Windows so we can build to pvtr.exe and skip bash-only steps.
# On Windows make (e.g. via choco/scoop), the OS env var is "Windows_NT".
ifeq ($(OS),Windows_NT)
	BINARY_NAME=pvtr.exe
	HAVE_BASH=$(shell where bash 2>NUL)
else
	BINARY_NAME=pvtr
	HAVE_BASH=$(shell command -v bash 2>/dev/null)
endif

.PHONY: build test testcov release binary tidy test-cov release-candidate release-nix release-nix-amd64 release-nix-arm64 release-win release-win-amd64 release-win-arm64 release-mac release-mac-amd64 release-mac-arm64 todo

build: tidy test binary
testcov: test test-cov
release: tidy test release-nix release-win release-mac

binary:
	@echo "  >  Building binary ..."
	go build -o $(BINARY_NAME) -ldflags="$(BUILD_FLAGS)"

test:
	@echo "  >  Validating code ..."
ifeq ($(HAVE_BASH),)
	@echo "  >  bash not found; skipping install_test.sh (Go tests still run)"
else
	@bash ./test/install_test.sh
endif
	@go vet ./...
	@go test ./...

tidy:
	@echo "  >  Tidying go.mod ..."
	@go mod tidy

test-cov:
	@echo "Running tests and generating coverage output ..."
	@go test ./... -coverprofile coverage.out -covermode count
	@echo "Current test coverage : $$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+') %"

release-candidate: tidy test
	@echo "  >  Building release candidate for Linux (amd64)..."
	$(BUILD_LINUX_AMD64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=nix-rc'"
	@echo "  >  Building release candidate for Linux (arm64)..."
	$(BUILD_LINUX_ARM64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=nix-rc'"
	@echo "  >  Building release candidate for Windows (amd64)..."
	$(BUILD_WIN_AMD64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=win-rc'"
	@echo "  >  Building release candidate for Windows (arm64)..."
	$(BUILD_WIN_ARM64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=win-rc'"
	@echo "  >  Building release candidate for Darwin (amd64)..."
	$(BUILD_MAC_AMD64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=darwin'"
	@echo "  >  Building release candidate for Darwin (arm64)..."
	$(BUILD_MAC_ARM64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=darwin'"

release-nix: release-nix-amd64 release-nix-arm64

release-nix-amd64:
	@echo "  >  Building release for Linux (amd64)..."
	$(BUILD_LINUX_AMD64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=linux'"

release-nix-arm64:
	@echo "  >  Building release for Linux (arm64)..."
	$(BUILD_LINUX_ARM64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=linux'"

release-win: release-win-amd64 release-win-arm64

release-win-amd64:
	@echo "  >  Building release for Windows (amd64)..."
	$(BUILD_WIN_AMD64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=windows'"

release-win-arm64:
	@echo "  >  Building release for Windows (arm64)..."
	$(BUILD_WIN_ARM64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=windows'"

release-mac: release-mac-amd64 release-mac-arm64

release-mac-amd64:
	@echo "  >  Building release for Darwin (amd64)..."
	$(BUILD_MAC_AMD64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=darwin'"

release-mac-arm64:
	@echo "  >  Building release for Darwin (arm64)..."
	$(BUILD_MAC_ARM64) -ldflags="$(BUILD_FLAGS) -X 'main.VersionPostfix=darwin'"

todo:
	@read -p "Write your todo here: " TODO; \
	echo "- [ ] $$TODO" >> TODO.md
