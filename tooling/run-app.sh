#!/bin/bash

set -e
set -o pipefail

if [[ "$1" =~ ^(--clean|-c)$ ]]; then
  echo "Running clean build (deleting volumes)"
  sudo docker compose down -v
fi

if docker ps | grep kafka; then
  echo "Kafka is running, checking if topics exist"
  if ! ./tooling/create-topic.sh --check; then
    echo "Topics do not exist, creating them"
    chmod +x ./tooling/create-topic.sh
    ./tooling/create-topic.sh --start
  else
    echo "Topics exist, starting app"
  fi
fi

# Build and start the application
if [[ "$2" =~ ^(--detach|-d)$ ]]; then
  echo "Starting application in detached mode..."
  sudo docker compose up --build -d
else
  echo "Starting application in foreground mode..."
  sudo docker compose up --build
fi