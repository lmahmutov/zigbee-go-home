VERSION  := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS  := -s -w -X main.version=$(VERSION)
BINARY   := zigbee-home
CMD      := ./cmd/zigbee-home

.PHONY: build build-linux-arm64 build-linux-mipsle build-linux-mips build-linux-arm build-minimal test vet clean

build:
	CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY) $(CMD)

build-linux-arm64:
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-arm64 $(CMD)

build-linux-mipsle:
	GOOS=linux GOARCH=mipsle GOMIPS=softfloat CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-mipsle $(CMD)

build-linux-mips:
	GOOS=linux GOARCH=mips GOMIPS=softfloat CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-mips $(CMD)

build-linux-arm:
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o $(BINARY)-linux-arm $(CMD)

build-minimal:
	CGO_ENABLED=0 go build -tags no_mqtt,no_automation -ldflags="$(LDFLAGS)" -o $(BINARY)-minimal $(CMD)

test:
	go test ./...

vet:
	go vet ./...

clean:
	rm -f $(BINARY) $(BINARY)-linux-* $(BINARY)-minimal
