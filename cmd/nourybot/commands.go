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

	// cmdParams are additional command parameters.
	// e.g. `()weather san antonio`
	// cmdParam[0] is `san` and cmdParam[1] = `antonio`.
	//
	// Since Twitch messages are at most 500 characters I use a
	// maximum count of 500+10 just to be safe.
	// https://discuss.dev.twitch.tv/t/missing-client-side-message-length-check/21316
	cmdParams := strings.SplitN(message.Message, " ", 500)

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
		"cmdParams", cmdParams,
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
	case "":
		if msgLen == 1 {
			reply = "xd"
		}
		// --------------------------------
		// pleb commands
		// --------------------------------
	case "gpt":
		if msgLen < 2 {
			reply = "Not enough arguments provided. Usage: ()bttv <emote name>"
		} else {
			app.chatPersonalContext(target, message.User.Name, message.Message[6:len(message.Message)])
		}

		if reply != "" {
			go app.Send(target, reply)
			return
		}

	}
}
