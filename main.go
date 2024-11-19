package main

import (
	"os"
	"strings"
	"fmt"
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
	twitchBotName  string // Bot name (e.g., "gpt")
	trigger        string // Command trigger (e.g., "()")
}

type application struct {
	twitchClient *twitch.Client
	log          *zap.SugaredLogger
	cfg          config
	userMsgStore map[string][]ollamaMessage // Use ollamaMessage
	msgStore     []ollamaMessage            // Use ollamaMessage
}

func main() {
	// Setup logger
	logger := zap.NewExample()
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Sugar().Errorw("error syncing logger", "error", err)
		}
	}()
	sugar := logger.Sugar()

	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		sugar.Fatal("Error loading .env")
	}

	// Initialize user message store and application
	userMsgStore := make(map[string][]ollamaMessage) // ollamaMessage is used here

	app := &application{
		log:          sugar,
		userMsgStore: userMsgStore,
	}

	// Load environment values into the application config
	app.cfg.twitchUsername = os.Getenv("TWITCH_USERNAME")
	app.cfg.twitchOauth = os.Getenv("TWITCH_OAUTH")
	app.cfg.ollamaModel = os.Getenv("OLLAMA_MODEL")
	app.cfg.ollamaContext = os.Getenv("OLLAMA_CONTEXT")
	app.cfg.ollamaSystem = os.Getenv("OLLAMA_SYSTEM")
	app.cfg.ollamahost = os.Getenv("OLLAMA_HOST")
	app.cfg.twitchBotName = os.Getenv("TWITCHBOTNAME")

	// Default to "gpt" if not set
	if app.cfg.twitchBotName == "" {
		app.cfg.twitchBotName = "gpt"
	}

	app.cfg.trigger = os.Getenv("TRIGGER")
	if app.cfg.trigger == "" {
		app.cfg.trigger = "()" // Default to "()" if not set
	}

	// Create a new Twitch client
	tc := twitch.NewClient(app.cfg.twitchUsername, app.cfg.twitchOauth)
	app.twitchClient = tc

	// On Private Message received
	app.twitchClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
		// Extract the roomId (channel)
		roomId := message.Tags["room-id"]
		if roomId == "" {
			return
		}

		// Check if the message starts with the trigger from the config.
		if len(message.Message) >= len(app.cfg.trigger) && message.Message[:len(app.cfg.trigger)] == app.cfg.trigger {
			var reply string

			// Extract the command name and arguments
			commandName := strings.ToLower(strings.SplitN(message.Message, " ", 3)[0][len(app.cfg.trigger):])
			args := strings.TrimSpace(message.Message[len(app.cfg.trigger)+len(commandName):])

			// Handle the dynamic command (based on TWITCHBOTNAME)
			if commandName == app.cfg.twitchBotName { // Use the dynamic bot name here
				if len(args) < 1 {
					reply = fmt.Sprintf("Not enough arguments provided. Usage: %s%s <query>", app.cfg.trigger, app.cfg.twitchBotName)
				} else {
					// Check the context configuration and process accordingly
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
			} else {
				reply = fmt.Sprintf("Unknown command: %s%s", app.cfg.trigger, commandName)
			}

			// Send the reply if needed
			if reply != "" {
				go app.send(message.Channel, reply)
				return
			}
		}
	})

	// Log successful connection
	app.twitchClient.OnConnect(func() {
		app.log.Info("Successfully connected to Twitch Servers")
		app.log.Infow("Ollama", "Context:", app.cfg.ollamaContext, "Model:", app.cfg.ollamaModel, "System:", app.cfg.ollamaSystem)
	})

	// Join specified channels
	channels := os.Getenv("TWITCH_CHANNELS")
	channel := strings.Split(channels, ",")
	for _, ch := range channel {
		app.twitchClient.Join(ch)
		app.twitchClient.Say(ch, "MrDestructoid")
		app.log.Infof("Joining channel: %s", ch)
	}

	// Connect to Twitch chat
	err = app.twitchClient.Connect()
	if err != nil {
		panic(err)
	}
}
