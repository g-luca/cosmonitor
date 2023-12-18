#!/bin/bash

# start the docker compose
docker compose -f compose-multi.yml down
sudo docker system prune --volumes -a -f