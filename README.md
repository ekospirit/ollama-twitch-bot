# NourybotGPT

Twitch chat bot that interacts with ollama. Work in Progress.

## Requirements:
Go

[Ollama.com](https://ollama.com)

## Build and run:
1. Change the default values in the provided `env.example` and rename it to `.env`.
2. Make sure ollama is up and running on the host.
3. With docker compose:
* `$ make up`
3. Without docker:
* `$ make build && make run`
4. Join the Twitch channel you chose and use `()gpt <cool query>` 
