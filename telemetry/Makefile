.PHONY: build build-win build-linux build-macos run clean install test vet

BINARY=mvgn-telemetry
BINARY_WIN=$(BINARY).exe

build:
	go build -o $(BINARY) ./cmd/mvgn-telemetry

build-win:
	GOOS=windows GOARCH=amd64 go build -o $(BINARY_WIN) ./cmd/mvgn-telemetry

build-linux:
	GOOS=linux GOARCH=amd64 go build -o $(BINARY)-linux ./cmd/mvgn-telemetry

build-macos:
	GOOS=darwin GOARCH=arm64 go build -o $(BINARY)-macos ./cmd/mvgn-telemetry

run: build
	./$(BINARY)

clean:
	rm -f $(BINARY) $(BINARY_WIN) $(BINARY)-linux $(BINARY)-macos

install: build
	cp $(BINARY) $(GOPATH)/bin/

test:
	go test ./...

vet:
	go vet ./...
