NAME=dirsearch
SOURCE=cmd/dirsearch/main.go

GOBUILD=go build

DEPEND=github.com/Masterminds/glide

# Command to get glide, you need to run it only once
.PHONY: get_glide
get_glide:
	go get -u -v $(DEPEND)
	$(GOPATH)/bin/glide install

# Command to install dependencies using glide
.PHONY: install_dependencies
install_dependencies:
	glide install

# Run tests in verbose mode with race detector and display coverage
.PHONY: test
test:
	go test -v -cover -race $(shell glide novendor)

# Removing artifacts
.PHONY: clean
clean:
	$(info * Cleaning build folder)
	@rm -rf build/*

# Building linux binaries
.PHONY: _build_linux
_build_linux:
	$(info * Building executable for linux x64 [$(SOURCE) -> build/linux_x64/$(NAME)])
	@GOOS=linux GOARCH=amd64 $(GOBUILD) -o build/linux_x64/$(NAME) $(SOURCE)

# Building osx binaries
.PHONY: _build_osx
_build_osx:
	$(info * Building executable for osx x64 [$(SOURCE) -> build/darwin_amd64/$(NAME)])
	@GOOS=darwin GOARCH=amd64 $(GOBUILD) -o build/darwin_amd64/$(NAME) $(SOURCE)

# Clean the build folder and then build executable for linux and osx
.PHONY: build
build: clean _build_linux _build_osx

# Run the application
.PHONY: run
run:
	go run cmd/dirsearch/main.go
