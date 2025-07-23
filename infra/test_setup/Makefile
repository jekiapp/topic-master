KAFKA_SERVICES = kafka kafka-ui kafka-publisher kafka-consumer
COMPOSE_FILE = docker-compose.yml

.PHONY: kafka-up kafka-down kafka-status

kafka-up:
	docker compose -f $(COMPOSE_FILE) up $(KAFKA_SERVICES)

kafka-stop:
	docker compose -f $(COMPOSE_FILE) stop $(KAFKA_SERVICES)

