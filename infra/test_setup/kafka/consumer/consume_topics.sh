#!/bin/bash

BROKER="kafka:9092"

# Wait for Kafka to be ready
until kafka-topics.sh --bootstrap-server $BROKER --list > /dev/null 2>&1; do
  echo "Waiting for Kafka broker..."
  sleep 2
done

echo "Kafka broker is ready."

# Read topics from file
mapfile -t topics < topics.txt

# Start a consumer for each topic in the background
for topic in "${topics[@]}"; do
  echo "Starting consumer for topic: $topic"
  kafka-console-consumer.sh --bootstrap-server $BROKER --topic "$topic" --group "test-group" --from-beginning &
done

# Wait for all background consumers
wait 