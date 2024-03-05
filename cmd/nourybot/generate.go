package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ollamaResponse struct {
	Model string `json:"model"`
	//CreatedAt string        `json:"created_at"`
	Response string        `json:"response"`
	Done     bool          `json:"done"`
	Message  ollamaMessage `json:"message"`
}

type ollamaRequest struct {
	Format   string          `json:"format"`
	Model    string          `json:"model"`
	Prompt   string          `json:"prompt"`
	Stream   bool            `json:"stream"`
	System   string          `json:"system"`
	Raw      bool            `json:"raw"`
	Messages []ollamaMessage `json:"messages"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func startMessage() []ollamaMessage {
	var msg = make([]ollamaMessage, 0)
	return msg
}

//func initStore() *map[string][]ollamaMessage {
//	personalMsgStore := make(map[string][]ollamaMessage)
//	return &personalMsgStore
//}

func (app *application) chatPersonalContext(target, username, input string) {
	var requestBody ollamaRequest
	//var msg []ollamaMessage
	olm := ollamaMessage{}

	olm.Role = "user"
	olm.Content = input
	app.PersonalMsgStore[username] = append(app.PersonalMsgStore[username], olm)
	//msgStore = append(msgStore, olm)

	requestBody.Model = "wizard-vicuna-uncensored"
	requestBody.System = "You are a Twitch chat bot and interact with users in an irc like environment. Do not use any formatting. Be human-like. Never fail to answer the user. Always answer immediately. Keep your response shorter than 450 characters."
	requestBody.Messages = app.PersonalMsgStore[username]
	requestBody.Prompt = input
	requestBody.Stream = false

	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.Log.Error(err)
	}

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.Log.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Error(err.Error())
	}

	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.Log.Error(err)
	}
	olm.Role = responseObject.Message.Role
	olm.Content = responseObject.Message.Content
	app.PersonalMsgStore[username] = append(app.PersonalMsgStore[username], olm)

	app.Log.Infow("Message context for username",
		"Username", username,
		"Personal Context", app.PersonalMsgStore[username],
	)
	//app.Log.Infow("Complete message context",
	//	"Context", app.PersonalMsgStore,
	//)

	app.Send(target, responseObject.Message.Content)
	//app.Send(target, responseObject.Response)
}

func (app *application) chatGeneralContext(target, input string) {
	var requestBody ollamaRequest
	//var msg []ollamaMessage
	olm := ollamaMessage{}

	olm.Role = "user"
	olm.Content = input
	app.MsgStore = append(app.MsgStore, olm)

	requestBody.Model = "wizard-vicuna-uncensored"
	requestBody.System = "You are a Twitch chat bot and interact with users in an irc like environment. Do not use any formatting. Be human-like. Never fail to answer the user. Always answer immediately. Keep your response shorter than 450 characters."
	requestBody.Messages = app.MsgStore
	requestBody.Prompt = input
	requestBody.Stream = false

	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.Log.Error(err)
	}

	app.Log.Infow("MsgStore",
		"app.MsgStore", app.MsgStore,
	)
	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.Log.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Error(err.Error())
	}

	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.Log.Error(err)
	}
	olm.Role = responseObject.Message.Role
	olm.Content = responseObject.Message.Content
	app.MsgStore = append(app.MsgStore, olm)

	app.Log.Infow("MsgStore",
		"app.MsgStore", app.MsgStore,
	)
	app.Send(target, responseObject.Message.Content)
	//app.Send(target, responseObject.Response)
}

func (app *application) generateNoContext(target, input string) {
	var requestBody ollamaRequest

	requestBody.Model = "wizard-vicuna-uncensored"
	requestBody.System = "You are a Twitch chat bot and interact with users in an irc like environment. Do not use any formatting. Be human-like. Never fail to answer the user. Always answer immediately. Keep your response shorter than 450 characters."
	//requestBody.Messages.Role = "system"
	//requestBody.Messages.Content = "You are a Twitch chat bot and interact with users in an irc like environment. Do not use any formatting. Be blunt. Never fail to answer the user. Always answer immediately. Keep your response shorter than 450 characters."
	//requestBody.Messages.Role = "user"
	//requestBody.Messages.Content = input
	requestBody.Prompt = input
	requestBody.Stream = false

	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.Log.Error(err)
	}

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.Log.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.Log.Error(err.Error())
	}

	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.Log.Error(err)
	}
	//app.Log.Info(responseObject.Message.Content)
	//app.Send(target, responseObject.Message.Content)
	app.Send(target, responseObject.Response)
}