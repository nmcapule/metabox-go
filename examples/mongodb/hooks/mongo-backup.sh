#!/bin/bash

rm -rf ./workspace

mongodump --out ./workspace \
    --host localhost \
    --port 27018 \
    --db small_default \
    --username simcel \
    --password simcel \
    --authenticationDatabase "admin"
