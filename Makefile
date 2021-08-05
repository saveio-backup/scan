GOFMT=gofmt
GC=go build --tags json1
LAST_VERSION := $(shell git describe --match "v*" --always --tags)
VERSION_PARTS      := $(subst ., ,$(LAST_VERSION))
MAJOR              := $(word 1,$(VERSION_PARTS))
MINOR              := $(word 2,$(VERSION_PARTS))
MICRO              := $(word 3,$(VERSION_PARTS))
NEXT_MICRO         := $(shell echo $$(($(MICRO)+1)))
VERSION := $(MAJOR).$(MINOR).$(NEXT_MICRO)
PYLONS_GITCOMMIT=$(shell cd .. && cd pylons && git log -1 --pretty=format:"%H")
CARRIER_GITCOMMIT=$(shell cd .. && cd carrier && git log -1 --pretty=format:"%H")
DSP_GITCOMMIT=$(shell cd .. && cd dsp-go-sdk && git log -1 --pretty=format:"%H")

BUILD_SCAN_PAR = -ldflags "-X github.com/saveio/scan/common/config.VERSION=$(VERSION)  -X github.com/saveio/pylons.Version=${PYLONS_GITCOMMIT} -X github.com/saveio/carrier/network.Version=${CARRIER_GITCOMMIT}  -X github.com/saveio/dsp-go-sdk.Version=${DSP_GITCOMMIT}"

client: clean 
	$(GC) $(BUILD_SCAN_PAR)  -o scan ./bin/scan/main.go


all: wddns lddns mdns

wscan:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GC) $(BUILD_SCAN_PAR) -o wscan.exe ./bin/scan/main.go

lscan:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GC) $(BUILD_SCAN_PAR) -o lscan ./bin/scan/main.go

mscan:
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GC) $(BUILD_SCAN_PAR) -o mscan ./bin/scan/main.go

format:
	$(GOFMT) -w main.go

clean:
	rm -rf *.8 *.o *.out *.6 *exe
	rm -f scan 
