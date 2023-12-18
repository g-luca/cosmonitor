#!/bin/bash
./stop.sh
# start the docker compose
docker compose -f compose-multi.yml up -d