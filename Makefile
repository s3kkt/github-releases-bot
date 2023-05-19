build:
	go build -o github-releases-bot cmd/bot/main.go

migrate:
	go run cmd/bot/main.go -config="./configs/config.yml" -migrations=true

run:
	go run cmd/bot/main.go -config="./configs/config.yml"

run-env:
	go run cmd/bot/main.go