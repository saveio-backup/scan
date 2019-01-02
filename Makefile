GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --always --tags)
BUILD_SEEDS_PAR = -ldflags "-X github.com/oniio/oniDNS/config/config.VERSION=$(VERSION)"

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)

seeds: $(SRC_FILES)
	$(GC)  $(BUILD_SEEDS_PAR) -o seeds main.go


all: wseeds lseeds dseeds

wseeds:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_SEEDS_PAR) -o wseeds.exe main.go

lseeds:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_SEEDS_PAR) -o lseeds main.go

dseeds:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_SEEDS_PAR) -o dseeds main.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe
	rm -rf do do-*
