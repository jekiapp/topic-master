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

# Create topics
for topic in "${topics[@]}"; do
  echo "Creating topic: $topic"
  kafka-topics.sh --bootstrap-server $BROKER --create --if-not-exists --topic "$topic" --partitions 1 --replication-factor 1
done

echo "All topics created successfully"

# Publish random messages to random topics in a loop
while true; do
  random_topic=${topics[$RANDOM % ${#topics[@]}]}
  timestamp=$(date +%s)
  message="{\"event\":\"$random_topic\",\"timestamp\":$timestamp,\"data\":{\"id\":$RANDOM}}"
  echo "Publishing to $random_topic: $message"
  echo "$message" | kafka-console-producer.sh --bootstrap-server $BROKER --topic "$random_topic" > /dev/null
  sleep 1
done 