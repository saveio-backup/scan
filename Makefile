GOFMT=gofmt
GC=go build
VERSION := $(shell git describe --abbrev=4 --always --tags)
BUILD_DDNS_PAR = -ldflags "-X github.com/saveio/scan/config/config.VERSION=$(VERSION)"

SRC_FILES = $(shell git ls-files | grep -e .go$ | grep -v _test.go)

ddns: $(SRC_FILES)
	$(GC)  $(BUILD_DDNS_PAR) -o ddns main.go


all: wddns lddns mdns

wddns:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_DDNS_PAR) -o wddns.exe main.go

lddns:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_DDNS_PAR) -o lddns main.go

mdns:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_DDNS_PAR) -o mdns main.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe
	rm -f ddns
