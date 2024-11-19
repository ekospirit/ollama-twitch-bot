package main

import (
	"os"
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

type config struct {
	twitchUsername string
	twitchOauth    string
	ollamaModel    string
	ollamaContext  string
	ollamaSystem   string
	ollamahost     string
}

type application struct {
	twitchClient *twitch.Client
	log          *zap.SugaredLogger
	cfg          config
	userMsgStore map[string][]ollamaMessage
	msgStore     []ollamaMessage
}

func main() {
	logger := zap.NewExample()
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Sugar().Errorw("error syncing logger",
				"error", err,
			)
		}
	}()
	sugar := logger.Sugar()

	err := godotenv.Load()
	if err != nil {
		sugar.Fatal("Error loading .env")
	}

	userMsgStore := make(map[string][]ollamaMessage)

	app := &application{
		log:          sugar,
		userMsgStore: userMsgStore,
	}

	app.cfg.twitchUsername = os.Getenv("TWITCH_USERNAME")
	app.cfg.twitchOauth = os.Getenv("TWITCH_OAUTH")
	app.cfg.ollamaModel = os.Getenv("OLLAMA_MODEL")
	app.cfg.ollamaContext = os.Getenv("OLLAMA_CONTEXT")
	app.cfg.ollamaSystem = os.Getenv("OLLAMA_SYSTEM")
	app.cfg.ollamahost = os.Getenv("OLLAMA_HOST")

	tc := twitch.NewClient(app.cfg.twitchUsername, app.cfg.twitchOauth)
	app.twitchClient = tc

	// Received a PrivateMessage (normal chat message).
	app.twitchClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// roomId is the Twitch UserID of the channel the message originated from.
		// If there is no roomId something went really wrong.
		roomId := message.Tags["room-id"]
		if roomId == "" {
			return
		}

		// Message was shorter than our prefix is therefore it's irrelevant for us.
		if len(message.Message) >= 2 && message.Message[:2] == "()" {
			var reply string

			// msgLen is the amount of words in a message without the prefix.
			msgLen := len(strings.SplitN(message.Message, " ", -2))

			// commandName is the actual name of the command without the prefix.
			// e.g. `()gpt` is `gpt`.
			commandName := strings.ToLower(strings.SplitN(message.Message, " ", 3)[0][2:])
			switch commandName {
			case "gpt":
				if msgLen < 2 {
					reply = "Not enough arguments provided. Usage: ()gpt <query>"
					return
				} else {
					switch app.cfg.ollamaContext {
					case "none":
						app.generateNoContext(message.Channel, message.Message[6:len(message.Message)])
						return

					case "general":
						app.chatGeneralContext(message.Channel, message.Message[6:len(message.Message)])
						return

					case "user":
						app.chatUserContext(message.Channel, message.User.Name, message.Message[6:len(message.Message)])
						return
					}
				}
				if reply != "" {
					go app.send(message.Channel, reply)
					return
				}
			}
		}
	})

	app.twitchClient.OnConnect(func() {
		app.log.Info("Successfully connected to Twitch Servers")
		app.log.Infow("Ollama",
			"Context: ", app.cfg.ollamaContext,
			"Model: ", app.cfg.ollamaModel,
			"System: ", app.cfg.ollamaSystem,
		)
	})

	channels := os.Getenv("TWITCH_CHANNELS")
	channel := strings.Split(channels, ",")
	for i := 0; i < len(channel); i++ {
		app.twitchClient.Join(channel[i])
		app.twitchClient.Say(channel[i], "MrDestructoid")
		app.log.Infof("Joining channel: %s", channel[i])
	}

	// Actually connect to chat.
	err = app.twitchClient.Connect()
	if err != nil {
		panic(err)
	}
}
