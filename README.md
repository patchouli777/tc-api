# What?

Pet project. Backend API server for twitch-like (livestreaming) app. Part of the twitch-clone project.

# Why?

Practice.

# Whats here?

* asynq task queue
* swagger docs
* redis
* docker
* sqlc
* grpc client

# How to run?
1. remove .example suffix for:
    1. secrets/postgres_password.example
    2. secrets/redis_password.example
    3. .env.example

2. for local dev run:
```sh
make local
make ssmock
make setup
make server
```
