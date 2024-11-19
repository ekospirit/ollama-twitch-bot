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
  twitchBotName  string
  trigger        string
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
  app.cfg.twitchBotName = os.Getenv("TWITCHBOTNAME")
  if app.cfg.twitchBotName == "" {
      app.cfg.twitchBotName = "gpt" // Fallback to default bot name
  }

  app.cfg.trigger = os.Getenv("TRIGGER")
  if app.cfg.trigger == "" {
      app.cfg.trigger = "()" // Fallback to default trigger
  }

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

    // Check if the message starts with the trigger from the config.
    if len(message.Message) >= len(app.cfg.trigger) && message.Message[:len(app.cfg.trigger)] == app.cfg.trigger {
        var reply string

        // Extract the command name and arguments.
        msgLen := len(strings.SplitN(message.Message, " ", -2))
        commandName := strings.ToLower(strings.SplitN(message.Message, " ", 3)[0][len(app.cfg.trigger):])
        args := strings.TrimSpace(message.Message[len(app.cfg.trigger)+len(commandName):])

        // Handle commands.
        switch commandName {
        case "gpt":
            if msgLen < 2 {
                reply = fmt.Sprintf("Not enough arguments provided. Usage: %sgpt <query>", app.cfg.trigger)
            } else {
                // Check the context configuration and process accordingly.
                switch app.cfg.ollamaContext {
                case "none":
                    app.generateNoContext(message.Channel, args)
                    return
                case "general":
                    app.chatGeneralContext(message.Channel, args)
                    return
                case "user":
                    app.chatUserContext(message.Channel, message.User.Name, args)
                    return
                default:
                    reply = "Invalid context configuration. Please check the server settings."
                }
            }
        default:
            reply = fmt.Sprintf("Unknown command: %s%s", app.cfg.trigger, commandName)
        }

        // Send the reply if there is one.
        if reply != "" {
            go app.send(message.Channel, reply)
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
