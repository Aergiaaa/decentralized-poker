build:
	@go build -o  bin/poker

build_m:
	@go build -o GOOS=android GOARCH=arm64 -o bin/poker.apk

run: build
	@./bin/poker

test: 
	@go test -v ./...