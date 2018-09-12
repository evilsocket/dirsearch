NAME=dirsearch
SOURCE=cmd/dirsearch/main.go

GOBUILD=go build

DEPEND=github.com/Masterminds/glide

ifeq ($(OS),Windows_NT)
    RM = rmdir /s /q build 2>nul
	BUILD_WIN = set GOOS=windows; set GOARCH=amd64
	BUILD_LINUX = set GOOS=linux; set GOARCH=amd64
	BUILD_OSX = set GOOS=darwin; set GOARCH=amd64
else
    RM = rm -rf build/*
	BUILD_WIN = GOOS=windows GOARCH=amd64
	BUILD_LINUX = GOOS=linux GOARCH=amd64
	BUILD_OSX = GOOS=darwin GOARCH=amd64
endif

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
	$(@shell if exist build $(RM))

# Building linux binaries
.PHONY: _build_linux
_build_linux:
	$(info * Building executable for linux x64 [$(SOURCE) -> build/linux_x64/$(NAME)])
	@ $(@shell BUILD_LINUX) $(GOBUILD) -o build/linux_x64/$(NAME) $(SOURCE)

# Building osx binaries
.PHONY: _build_osx
_build_osx:
	$(info * Building executable for osx x64 [$(SOURCE) -> build/darwin_amd64/$(NAME)])
	@ $(@shell BUILD_OSX) $(GOBUILD) -o build/darwin_amd64/$(NAME) $(SOURCE)

# Building windows binaries
.PHONY: _build_windows
_build_windows:
	$(info * Building executable for windows x64 [$(SOURCE) -> build/win64/$(NAME)])
	@ $(@shell BUILD_WIN) $(GOBUILD) -o build/win64/$(NAME).exe $(SOURCE)	

# Clean the build folder and then build executable for linux and osx
.PHONY: build
build: clean _build_windows _build_linux _build_osx

# Run the application
.PHONY: run
run:
	go run cmd/dirsearch/main.go
