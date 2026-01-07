.PHONY: up re fake local local-stop demo demo-stop swag test lint


swag:
	swag init \
	--parseInternal \
	-g ./internal/app/routes.go


up:
	docker-compose up


re:
	docker-compose up --build server


fake:
	go run .\cmd\fakeusers


local:
	docker run -d \
	--name twc-backend-postgres \
	-e POSTGRES_PASSWORD=123 \
	-e POSTGRES_DB=baklava \
	-e POSTGRES_USER=cherry \
	-p 5432:5432 \
	-v /var/lib/postgresql \
	--rm postgres:latest

	docker run -d \
	--name twc-backend-redis \
	-p 6379:6379 \
	--rm redis:latest


local-stop:
	docker stop twc-backend-postgres
	docker stop twc-backend-redis


demo:
	docker run -d \
	--name twc-backend-postgres \
	-e POSTGRES_PASSWORD=123 \
	-e POSTGRES_DB=baklava \
	-e POSTGRES_USER=cherry \
	-p 5432:5432 \
	-v /var/lib/postgresql \
	--rm postgres:latest

	docker run -d \
	--name twc-backend-redis \
	-p 6379:6379 \
	--rm redis:latest

	docker run -d \
	--name twc-streaming-server \
	-it \
	-p 1935:1935 \
	-p 1985:1985 \
	-p 8080:8080 \
	-p 8000:8000/udp \
	-p 10080:10080/udp \
	--rm ossrs/srs


demo-stop:
	docker stop twc-backend-postgres
	docker stop twc-backend-redis
	docker stop twc-streaming-server


test:
	go test ./tests


lint:
	golangci-lint run
