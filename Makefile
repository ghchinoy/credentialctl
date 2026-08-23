.PHONY: build test run clean tidy build-deps release-snapshot

# Default build binary
build:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o bin/credentialctl main.go

test:
	CGO_ENABLED=1 go test -v ./...

run:
	CGO_ENABLED=1 go run main.go

tidy:
	go mod tidy

# Build native library dependencies from credentio source if needed
build-deps:
	@echo "Checking native libcredentio_c build..."
	@if [ ! -f ../credentio-contributions/go/lib/libcredentio_c.dylib ] && [ ! -f ../credentio-contributions/go/lib/libcredentio_c.so ]; then \
		echo "Building native library via credentio-contributions scripts..."; \
		cd ../credentio-contributions && ./scripts/build-shared-lib.sh; \
	fi

# GoReleaser snapshot build
release-snapshot: build-deps
	goreleaser release --snapshot --clean

clean:
	rm -rf bin/ dist/
