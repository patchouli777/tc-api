# What?

Pet project. Backend API server for twitch-like (livestreaming) app. Part of the twitch-clone project.

# Why?

Practice.

# Whats here?

* haproxy load balancer
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

2. for "production" run*:
```sh
make up
```

3. for local dev run*:
```sh
make local
go run ./cmd/server
```
\* to populate with data run:
```
go run ./cmd/setup
```
