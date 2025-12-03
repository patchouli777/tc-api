#!/bin/sh
export POSTGRES_PASSWORD=$(cat /run/secrets/postgres_password)
# /myfolder/setup
/myfolder/server
