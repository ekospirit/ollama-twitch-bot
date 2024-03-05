package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type ollamaResponse struct {
	Model     string        `json:"model"`
	CreatedAt string        `json:"created_at"`
	Response  string        `json:"response"`
	Done      bool          `json:"done"`
	Message   ollamaMessage `json:"message"`
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

var requestBody ollamaRequest

// chatUserContext provides additional message context from specifically
// the past interactions of user with the AI since last restart.
func (app *application) chatUserContext(target, username, input string) {
	olm := ollamaMessage{}

	olm.Role = "user"
	olm.Content = input
	app.UserMsgStore[username] = append(app.UserMsgStore[username], olm)

	requestBody.Model = app.OllamaModel
	requestBody.System = app.OllamaSystem
	requestBody.Messages = app.UserMsgStore[username]
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
	app.UserMsgStore[username] = append(app.UserMsgStore[username], olm)

	app.Log.Infow("Message context for username",
		"Username", username,
		"Personal Context", app.UserMsgStore[username],
	)
	app.Send(target, responseObject.Message.Content)
}

// chatGeneralContext provides additional message context from every past
// interaction with the AI since last restart.
func (app *application) chatGeneralContext(target, input string) {
	olm := ollamaMessage{}

	olm.Role = "user"
	olm.Content = input
	app.MsgStore = append(app.MsgStore, olm)

	requestBody.Model = app.OllamaModel
	requestBody.System = app.OllamaSystem
	requestBody.Messages = app.MsgStore
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
	app.MsgStore = append(app.MsgStore, olm)

	app.Log.Infow("MsgStore",
		"app.MsgStore", app.MsgStore,
	)
	app.Send(target, responseObject.Message.Content)
}

// generateNoContext provides no additional message context
func (app *application) generateNoContext(target, input string) {
	var requestBody ollamaRequest

	requestBody.Model = app.OllamaModel
	requestBody.System = app.OllamaSystem
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

	app.Send(target, responseObject.Response)
}
