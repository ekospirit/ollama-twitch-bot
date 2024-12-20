# Ollama Twitch Bot

Twitch chat bot that interacts with ollama.

## Requirements:
[Golang](https://go.dev/)

[Ollama.com](https://ollama.com)

## Build and run:
1. Change the example values in the provided `env.example` and rename it to `.env`.
2. Make sure ollama is running on the host and reachable at `localhost:11434` and the model that you specified in the `.env` file is already downloaded and ready to go. (Can be checked with e.g. `ollama run wizard-vicuna-uncensored`)
3. Run:
    - With docker compose (might need sudo infront if you haven't setup rootless):
        - `$ make up`
    - Without docker:
        - `$ make build && make run`
4. Join the Twitch channels you chose and type `()gpt <cool query>` and hopefully get a response.

## Make Docker Image
1. Change the example values in the provided `env.example` and rename it to `.env`.
2. run `docker build -t twitchbot-ollama .`
3. Now you can run the image with `docker run -d --name twitchbot-ollama-container twitchbot-ollama`

## Example Response
![image](https://github.com/nouryxd/ollama-twitch-bot/assets/66651385/3a8a6e7d-07d7-42fc-bf10-27227746a1a8)
