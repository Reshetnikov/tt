#!/bin/sh

# Run from Dockerfile.prod
# or
# docker exec -it tt-app-1 scripts/migrate.sh up

if [ -z "$DB_USER" ] || [ -z "$DB_PASSWORD" ] || [ -z "$DB_HOST" ] || [ -z "$DB_PORT" ] || [ -z "$DB_NAME" ]; then
  echo "Not all required environment variables are set: DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME"
  exit 1
fi

DB_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

case "$1" in
  up)
    echo "Running migrations forward..."
    goose -dir db/migrations postgres "$DB_URL" up
    ;;
  down)
    echo "Rolling back migrations..."
    goose -dir db/migrations postgres "$DB_URL" down
    ;;
  redo)
    echo "Rolling back the last migration and reapplying it..."
    goose -dir db/migrations postgres "$DB_URL" redo
    ;;
  status)
    echo "Checking the status of migrations..."
    goose -dir db/migrations postgres "$DB_URL" status
    ;;
  *)
    echo "Usage: $0 {up|down|redo|status}"
    exit 1
    ;;
esac