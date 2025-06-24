#!/bin/bash

set -e
set -o pipefail

if [[ "$1" =~ ^(--clean|-c)$ ]]; then
  echo "Running clean build (deleting volumes)"
  sudo docker compose down -v
fi

sudo docker compose up --build