#!/bin/bash
set -e

MAX_RETRIES=30
RETRY_INTERVAL=1
CONTAINER_NAME="gophkeeper_db"

echo "Waiting for PostgreSQL container to become available..."

for i in $(seq 1 $MAX_RETRIES); do
if docker container inspect "$CONTAINER_NAME" --format '{{.State.Running}}' 2>/dev/null | grep -q "true"; then        echo "Database is ready!"
              echo "PostgreSQL container is running!"

    sleep 3

            echo "Database should be ready now!"
            exit 0
            fi
                echo "Waiting for PostgreSQL container to be ready... ($i/$MAX_RETRIES)"
    sleep $RETRY_INTERVAL

done

echo "PostgreSQL container failed to become ready in time"
exit 1