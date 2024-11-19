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

// chatUserContext provides additional message context from specifically
// the past interactions of user with the AI since last restart.
func (app *application) chatUserContext(target, username, input string) {
	// Prepare the user message
	olm := ollamaMessage{
		Role:    "user",
		Content: input,
	}
	app.userMsgStore[username] = append(app.userMsgStore[username], olm)

	// Prepare the system message if it is set
	var messages []ollamaMessage
	if app.cfg.ollamaSystem != "" {
		messages = append(messages, ollamaMessage{
			Role:    "system",
			Content: app.cfg.ollamaSystem,
		})
	}

	// Add the user messages to the context
	messages = append(messages, app.userMsgStore[username]...)

	// Prepare the request body with system and user messages
	requestBody := ollamaRequest{
		Model:    app.cfg.ollamaModel,
		System:   app.cfg.ollamaSystem,
		Messages: messages,
		Prompt:   input,
		Stream:   false,
	}

	// Marshal the request body
	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.log.Error("Error marshaling request body: ", err)
		return
	}

	// Use the ollamahost variable for the endpoint
	ollamahost := app.cfg.ollamahost

	// Send the request to the Ollama API
	resp, err := http.Post(ollamahost+"/api/chat", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.log.Error("Error sending request to Ollama: ", err.Error())
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.log.Error("Error reading response body: ", err.Error())
		return
	}

	// Parse the response
	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.log.Error("Error unmarshaling response body: ", err)
		return
	}

	// Ensure the response has valid content
	if responseObject.Message.Content == "" {
		app.log.Error("Received empty content in response from Ollama")
		return
	}

	// Create the response message object
	olm.Role = responseObject.Message.Role
	olm.Content = responseObject.Message.Content
	app.userMsgStore[username] = append(app.userMsgStore[username], olm)

	// Log the user message store for debugging
	app.log.Infow("Message context for username",
		"username", username,
		"app.userMsgStore[username]", app.userMsgStore[username],
	)

	// Send the response back to the target
	app.send(target, responseObject.Message.Content)
}


// chatGeneralContext provides additional message context from every past
// interaction with the AI since last restart.
func (app *application) chatGeneralContext(target, input string) {
	// Prepare the user message
	olm := ollamaMessage{
		Role:    "user",
		Content: input,
	}
	app.msgStore = append(app.msgStore, olm)

	// Prepare the system message if it is set
	var messages []ollamaMessage
	if app.cfg.ollamaSystem != "" {
		messages = append(messages, ollamaMessage{
			Role:    "system",
			Content: app.cfg.ollamaSystem,
		})
	}

	// Add the general message store to the context
	messages = append(messages, app.msgStore...)

	// Prepare the request body with system and user messages
	requestBody := ollamaRequest{
		Model:    app.cfg.ollamaModel,
		Messages: messages,
		Stream:   false,
	}

	// Marshal the request body
	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.log.Error("Error marshaling request body: ", err)
		return
	}

	// Use the ollamahost variable for the endpoint
	ollamahost := app.cfg.ollamahost

	// Send the request to the Ollama API
	resp, err := http.Post(ollamahost+"/api/chat", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.log.Error("Error sending request to Ollama: ", err.Error())
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.log.Error("Error reading response body: ", err.Error())
		return
	}

	// Parse the response
	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.log.Error("Error unmarshaling response body: ", err)
		return
	}

	// Ensure the response has valid content
	if responseObject.Message.Content == "" {
		app.log.Error("Received empty content in response from Ollama")
		return
	}

	// Create the response message object
	olm.Role = responseObject.Message.Role
	olm.Content = responseObject.Message.Content
	app.msgStore = append(app.msgStore, olm)

	// Log the message store for debugging
	app.log.Infow("app.msgStore",
		"app.msgStore", app.msgStore,
	)

	// Send the response back to the target
	app.send(target, responseObject.Message.Content)
}


// generateNoContext provides no additional message context
func (app *application) generateNoContext(target, input string) {
	// Prepare the request body with model, system, and user input
	requestBody := ollamaRequest{
		Model:   app.cfg.ollamaModel,
		System:  app.cfg.ollamaSystem,
		Prompt:  input,
		Stream:  false,
	}

	// Marshal the request body
	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.log.Error("Error marshaling request body: ", err)
		return
	}

	// Use the ollamahost variable for the endpoint
	ollamahost := app.cfg.ollamahost

	// Send the request to the Ollama API
	resp, err := http.Post(ollamahost+"/api/generate", "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.log.Error("Error sending request to Ollama: ", err.Error())
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.log.Error("Error reading response body: ", err.Error())
		return
	}

	// Parse the response
	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.log.Error("Error unmarshaling response body: ", err)
		return
	}

	// Ensure the response contains valid content
	if responseObject.Response == "" {
		app.log.Error("Received empty response content from Ollama")
		return
	}

	// Send the response back to the target
	app.send(target, responseObject.Response)
}
