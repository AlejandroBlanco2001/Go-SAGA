#!/bin/bash

set -e
set -o pipefail

if ! docker ps | grep kafka; then
  if [[ "$1" =~ ^(--start|-s)$ ]]; then
    echo "Kafka is not running, starting it..."
    docker compose up -d kafka
  else
    echo "Kafka is not running, not creating topic :("
    exit 1
  fi
fi

# Create the main topics for SAGA pattern events
docker exec -it kafka kafka-topics.sh --create --topic orders --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
docker exec -it kafka kafka-topics.sh --create --topic payments --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
docker exec -it kafka kafka-topics.sh --create --topic inventory --bootstrap-server localhost:9092 --partitions 3 --replication-factor 1
