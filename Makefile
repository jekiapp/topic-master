start-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml up

stop-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml down 
 
build-run:
	go build -o topic-master *.go
	./topic-master -data_path=infra/data/ -port=4181 -nsqlookupd_http_address=http://localhost:4161

build-linux-amd64:
	GOOS=linux GOARCH=amd64 go build -o bin/topic-master-linux-amd64 *.go

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -o bin/topic-master-linux-arm64 *.go

build-macos-amd64:
	GOOS=darwin GOARCH=amd64 go build -o bin/topic-master-darwin-amd64 *.go

build-macos-arm64:
	GOOS=darwin GOARCH=arm64 go build -o bin/topic-master-darwin-arm64 *.go

build-all: build-linux-amd64 build-linux-arm64 build-macos-amd64 build-macos-arm64