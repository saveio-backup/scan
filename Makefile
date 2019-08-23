GOFMT=gofmt
GC=go build
# VERSION := $(shell git describe --abbrev=4 --always --tags)
VERSION := $(shell git tag -l --sort=-v:refname | grep v1.0. | head -1)
BUILD_SCAN_PAR = -ldflags "-X github.com/saveio/scan/common/config.VERSION=$(VERSION)"

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)

client: clean 
	$(GC) $(BUILD_SCAN_PAR) --tags json1 -o scan ./bin/scan/main.go


all: wddns lddns mdns

wscan:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_SCAN_PAR) -o wscan.exe main.go

lscan:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_SCAN_PAR) -o lscan main.go

mscan:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_SCAN_PAR) -o mscan main.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe
	rm -f scan 
