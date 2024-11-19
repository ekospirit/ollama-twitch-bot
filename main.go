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
    twitchChannels string
    twitchBotName  string
    trigger        string
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
        sugar.Fatal("Error loading .env file")
    }

    userMsgStore := make(map[string][]ollamaMessage)

    app := &application{
        log:          sugar,
        userMsgStore: userMsgStore,
        cfg: config{
            twitchUsername: os.Getenv("TWITCH_USERNAME"),
            twitchOauth:    os.Getenv("TWITCH_OAUTH"),
            twitchChannels: os.Getenv("TWITCH_CHANNELS"),
            twitchBotName:  os.Getenv("TWITCHBOTNAME"),
            trigger:        os.Getenv("TRIGGER"),
            ollamaModel:    os.Getenv("OLLAMA_MODEL"),
            ollamaContext:  os.Getenv("OLLAMA_CONTEXT"),
            ollamaSystem:   os.Getenv("OLLAMA_SYSTEM"),
            ollamahost:     os.Getenv("OLLAMA_HOST"),
        },
    }

    tc := twitch.NewClient(app.cfg.twitchUsername, app.cfg.twitchOauth)
    app.twitchClient = tc

    // Received a PrivateMessage (normal chat message).
    app.twitchClient.OnPrivateMessage(func(message twitch.PrivateMessage) {
        roomId := message.Tags["room-id"]
        if roomId == "" {
            return
        }

        if len(message.Message) >= len(app.cfg.trigger) && message.Message[:len(app.cfg.trigger)] == app.cfg.trigger {
            var reply string
            msgLen := len(strings.SplitN(message.Message, " ", -2))
            commandName := strings.ToLower(strings.SplitN(message.Message, " ", 3)[0][len(app.cfg.trigger):])
            if commandName == strings.ToLower(app.cfg.twitchBotName) {
                if msgLen < 2 {
                    reply = "Not enough arguments provided. Usage: " + app.cfg.trigger + app.cfg.twitchBotName + " <query>"
                } else {
                    switch app.cfg.ollamaContext {
                    case "none":
                        app.generateNoContext(message.Channel, message.Message[len(app.cfg.trigger)+len(app.cfg.twitchBotName)+1:])
                    case "general":
                        app.chatGeneralContext(message.Channel, message.Message[len(app.cfg.trigger)+len(app.cfg.twitchBotName)+1:])
                    case "user":
                        app.chatUserContext(message.Channel, message.User.Name, message.Message[len(app.cfg.trigger)+len(app.cfg.twitchBotName)+1:])
                    }
                }
                if reply != "" {
                    go app.send(message.Channel, reply)
                }
            }
        }
    })

    app.twitchClient.OnConnect(func() {
        app.log.Info("Successfully connected to Twitch Servers")
        app.log.Infow("Ollama",
            "Context", app.cfg.ollamaContext,
            "Model", app.cfg.ollamaModel,
            "System", app.cfg.ollamaSystem,
        )
    })

    channels := strings.Split(app.cfg.twitchChannels, ",")
    for _, channel := range channels {
        app.twitchClient.Join(channel)
        app.twitchClient.Say(channel, "MrDestructoid")
        app.log.Infof("Joining channel: %s", channel)
    }

    err = app.twitchClient.Connect()
    if err != nil {
        sugar.Fatal(err)
    }
}
