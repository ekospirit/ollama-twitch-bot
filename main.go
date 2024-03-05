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

	//tc := twitch.NewClient(config.twitchUsername, config.twitchOauth)

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
		if len(message.Message) >= 2 {
			// Check if the first 2 characters of the mesage were our prefix.
			// if they were forward the message to the command handler.
			if message.Message[:2] == "()" {
				go app.handleCommand(message)
				return
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
