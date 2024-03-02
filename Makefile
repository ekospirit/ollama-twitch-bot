BINARY_NAME=Nourybot.out
BINARY_NAME_API=NourybotApi.out

up:
	docker compose up

down:
	docker compose down

rebuild:
	docker compose down
	docker compose build
	docker compose up -d

xd:
	docker compose down
	docker compose build
	docker compose up 
