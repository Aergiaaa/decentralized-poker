build:
	@go build -o  bin/poker.exe

build_m:
	@go build -o GOOS=android GOARCH=arm64 -o bin/poker.apk

run: build
	@./bin/poker.exe

test: 
	@go test -v ./...