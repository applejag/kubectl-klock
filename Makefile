ifeq ($(OS),Windows_NT)
BINARY := klock.exe
else
BINARY := klock
endif

.PHONY: build
build: bin/${BINARY}

bin/${BINARY}: bin cmd/*.go pkg/*/*.go
	go build -o bin/${BINARY}

bin:
	mkdir bin

.PHONY: clean
clean:
	rm -fv bin/${BINARY}

.PHONY: check
check:
	go test ./... -coverprofile cover.out

.PHONY: deps
deps: deps-go

.PHONY: deps-go
deps-go:
	go get
