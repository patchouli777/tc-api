.PHONY: up re fake local local-stop swag test


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
	-e POSTGRES_DB=twitchclone \
	-e POSTGRES_USER=less \
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


test:
	go test ./tests
