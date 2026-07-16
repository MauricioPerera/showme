package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// This file is the one exception in the repo whose forbids drops
// 'network' and 'llm': its entire purpose is calling an OpenAI-compatible
// HTTP endpoint that serves an LLM. generate-slide-content-usecase (the
// port) stays network-free by depending only on the ContentGenerator
// interface; OpenAIClient is the concrete adapter a caller injects there.
// Same reasoning as test-command-gate.md's exception for 'subprocess'.

const defaultMaxTokens = 512

// OpenAIClient calls an OpenAI-compatible chat completions endpoint (e.g.
// llama.cpp/LM Studio/vLLM servers) to generate slide content.
type OpenAIClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOpenAIClient builds an OpenAIClient targeting baseURL (e.g.
// "http://127.0.0.1:8080/v1") with the given model name.
func NewOpenAIClient(baseURL, model string) *OpenAIClient {
	return &OpenAIClient{
		baseURL:    baseURL,
		model:      model,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model              string          `json:"model"`
	Messages           []chatMessage   `json:"messages"`
	MaxTokens          int             `json:"max_tokens"`
	ChatTemplateKwargs map[string]bool `json:"chat_template_kwargs"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func buildPrompt(request GenerateContentRequest) string {
	prompt := "Intent: " + request.Intent
	if request.Context != "" {
		prompt += "\n\nContexto:\n" + request.Context
	}
	return prompt
}

// GenerateContent implements ContentGenerator by calling the chat
// completions endpoint and returning the assistant's message content.
func (c *OpenAIClient) GenerateContent(request GenerateContentRequest) (string, error) {
	reqBody := chatCompletionRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: "Sos un asistente que escribe el contenido de una diapositiva de una presentacion. Respondé solo con el texto de la diapositiva, sin explicaciones adicionales. Basate exclusivamente en el contexto dado; si el contexto no alcanza, decilo explicitamente en vez de inventar."},
			{Role: "user", Content: buildPrompt(request)},
		},
		MaxTokens:          defaultMaxTokens,
		ChatTemplateKwargs: map[string]bool{"enable_thinking": false},
	}

	encoded, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Post(c.baseURL+"/chat/completions", "application/json", bytes.NewReader(encoded))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ai provider returned status %d", resp.StatusCode)
	}

	var parsed chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("ai provider returned no choices")
	}

	return parsed.Choices[0].Message.Content, nil
}
