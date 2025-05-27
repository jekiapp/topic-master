start-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml up

stop-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml down 
 
build-run:
	go build -o topic-master *.go
	./topic-master -data_dir=infra/data -port=4181 -nsqlookupd_http_address=http://localhost:4161
