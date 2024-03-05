package main

import (
	"strings"

	"github.com/gempir/go-twitch-irc/v4"
)

// handleCommand takes in a twitch.PrivateMessage and then routes the message
// to the function that is responsible for each command and knows how to deal
// with it accordingly.
func (app *application) handleCommand(message twitch.PrivateMessage) {
	var reply string

	if message.Channel == "forsen" {
		return
	}

	// commandName is the actual name of the command without the prefix.
	// e.g. `()ping` would be `ping`.
	commandName := strings.ToLower(strings.SplitN(message.Message, " ", 3)[0][2:])

	// msgLen is the amount of words in a message without the prefix.
	// Useful to check if enough cmdParams are provided.
	msgLen := len(strings.SplitN(message.Message, " ", -2))

	// target is the channelname the message originated from and
	// where the TwitchClient should send the response
	target := message.Channel
	app.Log.Infow("Command received",
		// "message", message, // Pretty taxing
		"message.Message", message.Message,
		"message.Channel", target,
		"commandName", commandName,
		"msgLen", msgLen,
	)

	// A `commandName` is every message starting with `()`.
	// Hardcoded commands have a priority over database commands.
	// Switch over the commandName and see if there is a hardcoded case for it.
	// If there was no switch case satisfied, query the database if there is
	// a data.CommandModel.Name equal to the `commandName`
	// If there is return the data.CommandModel.Text entry.
	// Otherwise we ignore the message.
	switch commandName {
	case "gpt":
		if msgLen < 2 {
			reply = "Not enough arguments provided. Usage: ()gpt <query>"
		} else {
			//app.generateNoContext(target, message.User.Name, message.Message[6:len(message.Message)])
			//app.chatGeneralContext(target, message.User.Name, message.Message[6:len(message.Message)])
			app.chatUserContext(target, message.User.Name, message.Message[6:len(message.Message)])
		}

		if reply != "" {
			go app.Send(target, reply)
			return
		}
	}
}
