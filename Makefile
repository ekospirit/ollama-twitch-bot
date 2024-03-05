build:
	go get -d -v ./...
	go build -o OllamaTwitchBot.out .

run:
	./OllamaTwitchBot.out

jq:
	./OllamaTwitchBot.out | jq

up:
	docker compose down
	docker compose build
	docker compose up

rebuild:
	docker compose down
	docker compose build
	docker compose up -d

