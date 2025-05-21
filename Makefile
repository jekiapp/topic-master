start-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml up

stop-test-setup:
	docker compose -f infra/test_setup/docker-compose.yml down 
 