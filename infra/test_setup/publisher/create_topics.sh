#!/bin/bash

# Wait for nsqd to be ready
until curl -s http://nsqd:4151/ping > /dev/null; do
  echo "Waiting for nsqd..."
  sleep 1
done

# Read topics from file
mapfile -t topics < topics.txt

# Create topics
for topic in "${topics[@]}"; do
  echo "Creating topic: $topic"
  curl -X POST "http://nsqd:4151/topic/create?topic=$topic"
done

echo "All topics created successfully"

# after creating the topics, publish messages to random topics each second
while true; do
  # Select a random topic
  random_topic=${topics[$RANDOM % ${#topics[@]}]}
  
  # Generate a random message with timestamp
  timestamp=$(date +%s)
  message="{\"event\":\"$random_topic\",\"timestamp\":$timestamp,\"data\":{\"id\":$RANDOM}}"
  
  echo "Publishing to $random_topic: $message"
  curl -X POST "http://nsqd:4151/pub?topic=$random_topic" -d "$message"
  
  sleep 1
done 