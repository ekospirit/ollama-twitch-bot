package main

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

// handleCommand is called each time a message starts with "()"
func (app *application) handleCommand(message twitch.PrivateMessage) {
	var reply string

	// msgLen is the amount of words in a message without the prefix.
	// Useful to check if enough cmdParams are provided.
	msgLen := len(strings.SplitN(message.Message, " ", -2))

	// commandName is the actual name of the command without the prefix.
	// e.g. `()gpt` is `gpt`.
	commandName := strings.ToLower(strings.SplitN(message.Message, " ", 3)[0][2:])
	switch commandName {
	case "gpt":
		if msgLen < 2 {
			reply = "Not enough arguments provided. Usage: ()gpt <query>"
		} else {
			switch app.config.ollamaContext {
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
