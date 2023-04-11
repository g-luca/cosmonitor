#!/bin/bash

# start the docker compose
docker compose down
sudo docker system prune --volumes -a -f