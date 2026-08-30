.PHONY: build test run clean tidy fetch-credentio-lib release-snapshot docs-dev docs-build

# CGO compilation and linking flags for prebuilt Credentio C-ABI library
CREDENTIO_CFLAGS := -I$(PWD)/third_party/credentio/include

UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
CREDENTIO_LDFLAGS := -L$(PWD)/third_party/credentio/lib -lcredentio_c \
	-Wl,-rpath,@loader_path \
	-Wl,-rpath,@executable_path \
	-Wl,-rpath,@executable_path/../lib \
	-Wl,-rpath,@executable_path/lib \
	-Wl,-rpath,@loader_path/../third_party/credentio/lib \
	-Wl,-rpath,$(PWD)/third_party/credentio/lib
else
CREDENTIO_LDFLAGS := -L$(PWD)/third_party/credentio/lib -lcredentio_c \
	-Wl,-rpath,'$$ORIGIN' \
	-Wl,-rpath,'$$ORIGIN/../lib' \
	-Wl,-rpath,'$$ORIGIN/lib' \
	-Wl,-rpath,'$$ORIGIN/../third_party/credentio/lib' \
	-Wl,-rpath,$(PWD)/third_party/credentio/lib
endif

# Version and build metadata ldflags
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.5-dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -s -w -X github.com/ghchinoy/credentialctl/cmd.version=$(VERSION) -X github.com/ghchinoy/credentialctl/cmd.commit=$(COMMIT) -X github.com/ghchinoy/credentialctl/cmd.date=$(DATE)

# Default build binary
build: fetch-credentio-lib
	CGO_ENABLED=1 CGO_CFLAGS="$(CREDENTIO_CFLAGS) $(CGO_CFLAGS)" CGO_LDFLAGS="$(CREDENTIO_LDFLAGS) $(CGO_LDFLAGS)" go build -ldflags="$(LDFLAGS)" -o bin/credentialctl main.go

test: fetch-credentio-lib
	CGO_ENABLED=1 CGO_CFLAGS="$(CREDENTIO_CFLAGS) $(CGO_CFLAGS)" CGO_LDFLAGS="$(CREDENTIO_LDFLAGS) $(CGO_LDFLAGS)" go test -v ./...

run: fetch-credentio-lib
	CGO_ENABLED=1 CGO_CFLAGS="$(CREDENTIO_CFLAGS) $(CGO_CFLAGS)" CGO_LDFLAGS="$(CREDENTIO_LDFLAGS) $(CGO_LDFLAGS)" go run -ldflags="$(LDFLAGS)" main.go

tidy:
	go mod tidy

# Download prebuilt native library from GitHub Releases if missing
fetch-credentio-lib:
	@if [ ! -f third_party/credentio/lib/libcredentio_c.dylib ] && [ ! -f third_party/credentio/lib/libcredentio_c.so ]; then \
		echo "Fetching prebuilt native Credentio library..."; \
		./scripts/fetch-credentio-lib.sh; \
	fi

# GoReleaser snapshot build
release-snapshot: fetch-credentio-lib
	goreleaser release --snapshot --clean

# Documentation site
docs-dev:
	cd docs-site && npm run dev

docs-build:
	cd docs-site && npm run build

clean:
	rm -rf bin/ dist/ third_party/credentio/lib/ docs-site/dist/ docs-site/.astro/
