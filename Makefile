build:
	go get -d -v ./...
	go build -o Nourybot.out .

run:
	./Nourybot.out

jq:
	./Nourybot.out | jq


up:
	docker compose down
	docker compose build
	docker compose up

rebuild:
	docker compose down
	docker compose build
	docker compose up -d

