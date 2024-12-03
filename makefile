.PHONY: dev start build dist release clean

# Get the current Git tag
VERSION := $(shell git describe --tags --abbrev=0)

# Supported platforms and architectures
PLATFORMS := windows linux
ARCHS := amd64 386

dev: 
	air

start:
	go run main.go

# Build all distributions (uncompressed)
dist: $(foreach platform,$(PLATFORMS),$(foreach arch,$(ARCHS),dist-$(platform)-$(arch)))

dist-%:
	@mkdir -p dist/argus-$*-$(VERSION)
	GOOS=$(word 1,$(subst -, ,$*)) GOARCH=$(word 2,$(subst -, ,$*)) go build -o dist/argus-$*-$(VERSION)/argus$(if $(findstring windows,$*),.exe)
	cp argus-config.yml dist/argus-$*-$(VERSION)/argus-config.yml

# Release all compressed builds
release: $(foreach platform,$(PLATFORMS),$(foreach arch,$(ARCHS),release-$(platform)-$(arch)))

release-%: dist-%
	@mkdir -p releases
	tar -czf releases/argus-$*-$(VERSION).tar.gz -C dist argus-$*-$(VERSION)

# Clean build artifacts
clean:
	rm -rf dist
	rm -rf releases/*.tar.gz
