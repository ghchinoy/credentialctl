.PHONY: build test run clean tidy fetch-credentio-lib release-snapshot docs-dev docs-build

# CGO compilation and linking flags for prebuilt Credentio C-ABI library
CGO_CFLAGS ?= -I$(PWD)/third_party/credentio/include
CGO_LDFLAGS ?= -L$(PWD)/third_party/credentio/lib -lcredentio_c -Wl,-rpath,@loader_path -Wl,-rpath,$(PWD)/third_party/credentio/lib

# Default build binary
build: fetch-credentio-lib
	CGO_ENABLED=1 CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go build -ldflags="-s -w" -o bin/credentialctl main.go

test: fetch-credentio-lib
	CGO_ENABLED=1 CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go test -v ./...

run: fetch-credentio-lib
	CGO_ENABLED=1 CGO_CFLAGS="$(CGO_CFLAGS)" CGO_LDFLAGS="$(CGO_LDFLAGS)" go run main.go

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
