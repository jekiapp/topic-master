start-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml up

stop-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml down 
 
build-run:
	go build -o nsqper *.go
	./nsqper -data_dir=infra/data -port=4181 -nsqlookupd_http_address=http://localhost:4161
