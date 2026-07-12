#!/usr/bin/env bash
cd /home/lora/repos/ecomatrix/apps/backend
exec env ECOMATRIX_DB_DSN="postgres://repotwin:repotwin@localhost:5432/ecomatrix?sslmode=disable" \
  ECOMATRIX_DEV=true \
  ./bin/server
