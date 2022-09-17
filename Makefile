# Go parameters
GOCMD=go
MODULENAME=github.com/h44z/gmclogg
GOFILES:=$(shell go list ./... | grep -v /vendor/)
BUILDDIR=dist
BINARIES=$(subst cmd/,,$(wildcard cmd/*))

.PHONY: all test clean phony

all: dep

build: dep $(addsuffix -amd64,$(addprefix $(BUILDDIR)/,$(BINARIES)))

build-raspi:
	docker build -t raspibuilder ./build
	docker run -it --rm \
      -v $(shell pwd):/build \
      raspibuilder

build-cross-plat: dep $(addsuffix -arm64,$(addprefix $(BUILDDIR)/,$(BINARIES)))

dep:
	$(GOCMD) mod download

validate: dep
	$(GOCMD) fmt $(GOFILES)
	$(GOCMD) vet $(GOFILES)
	$(GOCMD) test -race $(GOFILES)

coverage: dep
	$(GOCMD) fmt $(GOFILES)
	$(GOCMD) test $(GOFILES) -v -coverprofile .testCoverage.txt
	$(GOCMD) tool cover -func=.testCoverage.txt  # use total:\s+\(statements\)\s+(\d+.\d+\%) as Gitlab CI regextotal:\s+\(statements\)\s+(\d+.\d+\%)

coverage-html: coverage
	$(GOCMD) tool cover -html=.testCoverage.txt

test: dep
	$(GOCMD) test $(MODULENAME)/... -v -count=1

clean:
	$(GOCMD) clean $(GOFILES)
	rm -rf .testCoverage.txt
	rm -rf $(BUILDDIR)

$(BUILDDIR)/%-amd64: cmd/%/main.go dep phony
	GOOS=linux GOARCH=amd64 $(GOCMD) build -o $@ $<

# Execute in Docker container
$(BUILDDIR)/%-arm64: cmd/%/main.go dep phony
	CGO_ENABLED=1 PKG_CONFIG_PATH=/usr/lib/aarch64-linux-gnu/pkgconfig \
	CC="zig cc -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu" \
	CXX="zig c++ -target aarch64-linux-gnu -isystem /usr/include -L/usr/lib/aarch64-linux-gnu" \
	GOOS=linux GOARCH=arm64 $(GOCMD) build -ldflags " -s -w -linkmode external" -o $@ $<
