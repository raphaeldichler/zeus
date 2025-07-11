GO_VERSION_FILE := .go-version

test:
	go test -v -cover ./...

build-grpc:
	find ./internal -name "*.proto" -print0 | \
		xargs -0 -n1 \
		protoc -I . \
			--go_out=. \
			--go_opt=paths=source_relative \
			--go-grpc_out=. \
			--go-grpc_opt=paths=source_relative

install-go:
	@set -e; \
	required_version=$$(cat $(GO_VERSION_FILE)); \
	if command -v go >/dev/null 2>&1; then \
		current_version=$$(go version | awk '{print $$3}' | sed 's/go//'); \
		ver_ge_121=$$(printf '%s\n' "$$current_version" "1.21.0" | sort -V | head -n1); \
		if [ "$$current_version" = "$$required_version" ]; then \
			echo "Go $$required_version is already installed."; \
			exit 0; \
		elif [ "$$ver_ge_121" = "1.21.0" ]; then \
			echo "Go $$current_version is >= 1.21.0, using toolchain support. Skipping install. (note export GOTOOLCHAIN=auto)"; \
			exit 0; \
		fi; \
	fi; \
	echo "Installing Go $$required_version..."; \
	arch=$$(uname -m); \
	case "$$arch" in \
		x86_64) goarch="amd64";; \
		aarch64) goarch="arm64";; \
		*) echo "Unsupported architecture: $$arch"; exit 1;; \
	esac; \
	url="https://go.dev/dl/go$${required_version}.linux-$${goarch}.tar.gz"; \
	echo "Downloading $$url ..."; \
	curl -LO "$$url"; \
	echo "Extracting..."; \
	sudo rm -rf /usr/local/go; \
	sudo tar -C /usr/local -xzf go$${required_version}.linux-$${goarch}.tar.gz; \
	rm go$${required_version}.linux-$${goarch}.tar.gz; \
	echo "Go $$required_version installed successfully.";

install-dependencies:
	$(MAKE) install-go
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	export PATH="$PATH:$(go env GOPATH)/bin"
	$(MAKE) build-grpc
