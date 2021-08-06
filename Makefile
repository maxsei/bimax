GOOS_GOARCH = darwin/386 darwin/amd64 linux/386 linux/amd64 linux/arm linux/arm64	windows/386 windows/amd64 windows/arm

.PHONY: shared
releases = $(GOOS_GOARCH)

shared:
	go build -v -x -buildmode=c-shared -o libbimax.so c/bimax.go
