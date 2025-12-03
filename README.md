# What?

Pet project. Backend API server for twitch-like (livestreaming) app. Part of the \<TODO: insert repo\>

# Why?

Practice.

# Whats here?

* haproxy load balancer
* swagger docs
* postgres 2 instances: 1 read, 1 write
* redis
* docker
* sqlc

# How to run?
1. remove .example suffix for:
    1. secrets/postgres_password.example
    2. secrets/redis_password.example
    3. .env.example

2. for "production" run*:
```sh
make dockup
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

\* to simulate activity run:
```sh
go run ./cmd/fakeusers
```

# License
))
