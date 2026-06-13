DB_DSN=postgres://postgres:postgres@localhost:5433/ai_tutor_bot?sslmode=disable

run:
	go run ./cmd/bot

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

migrate-status:
	go run ./cmd/migrate status

docker-up:
	docker compose up -d

docker-down:
	docker compose down