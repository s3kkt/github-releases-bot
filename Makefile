build:
	go build -o github-releases-bot cmd/bot/main.go

run:
	go run cmd/bot/main.go -config="./configs/config.yml"