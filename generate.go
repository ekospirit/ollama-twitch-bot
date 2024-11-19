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
	app.userMsgStore[username] = append(app.userMsgStore[username], olm)

	// Prepare the system message if it is set
	var messages []ollamaMessage

	// Include system message as the first message if it exists
	if app.cfg.ollamaSystem != "" {
		systemMessage := ollamaMessage{
			Role:    "system",
			Content: app.cfg.ollamaSystem,
		}
		messages = append(messages, systemMessage)
	}

	// Add the user messages to the context
	messages = append(messages, app.userMsgStore[username]...)

	// Prepare the request body with system and user messages
	requestBody := struct {
		Model   string        `json:"model"`
		Messages []ollamaMessage `json:"messages"`
		Stream  bool          `json:"stream"`
	}{
		Model:   app.cfg.ollamaModel,
		Messages: messages,
		Stream:  false,
	}

	// Marshal the request body
	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.log.Error(err)
		return
	}

	// Send the request to the Ollama API
  endpoint := fmt.Sprintf("%s/api/generate", app.cfg.ollamahost)

	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.log.Error(err.Error())
		return
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.log.Error(err.Error())
		return
	}

	// Parse the response
	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.log.Error(err)
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
	olm := ollamaMessage{}

	olm.Role = "user"
	olm.Content = input
	app.msgStore = append(app.msgStore, olm)

	// Prepare the system message if it is set
	var messages []ollamaMessage

	// Include system message as the first message if it exists
	if app.cfg.ollamaSystem != "" {
		systemMessage := ollamaMessage{
			Role:    "system",
			Content: app.cfg.ollamaSystem,
		}
		messages = append(messages, systemMessage)
	}

	// Add the general message store to the context
	messages = append(messages, app.msgStore...)

	// Prepare the request body with system and user messages
	requestBody := struct {
		Model    string        `json:"model"`
		Messages []ollamaMessage `json:"messages"`
		Stream   bool          `json:"stream"`
	}{
		Model:    app.cfg.ollamaModel,
		Messages: messages,
		Stream:   false,
	}

	// Marshal the request body
	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.log.Error(err)
		return
	}

	// Use the configured endpoint
	endpoint := fmt.Sprintf("%s/api/chat", app.cfg.ollamahost)

	// Send the request to the Ollama API
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.log.Error(err.Error())
		return
	}

	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.log.Error(err.Error())
		return
	}

	// Parse the response
	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.log.Error(err)
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
	var requestBody ollamaRequest

	requestBody.Model = app.cfg.ollamaModel
	requestBody.System = app.cfg.ollamaSystem
	requestBody.Prompt = input
	requestBody.Stream = false

	marshalled, err := json.Marshal(requestBody)
	if err != nil {
		app.log.Error(err)
	}

  endpoint := fmt.Sprintf("%s/api/generate", app.cfg.ollamahost)
  
	resp, err := http.Post(endpoint, "application/json", bytes.NewBuffer(marshalled))
	if err != nil {
		app.log.Error(err.Error())
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		app.log.Error(err.Error())
	}

	var responseObject ollamaResponse
	if err := json.Unmarshal(body, &responseObject); err != nil {
		app.log.Error(err)
	}

	app.send(target, responseObject.Response)
}
