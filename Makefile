build:
	go get -d -v ./...
	go build -o Nourybot.out .

run:
	./Nourybot.out

up:
	docker compose down
	docker compose build
	docker compose up

rebuild:
	docker compose down
	docker compose build
	docker compose up -d

