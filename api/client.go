package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/honganh1206/clue/message"
	"github.com/honganh1206/clue/server/conversation"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11435"
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) CreateConversation() (*conversation.Conversation, error) {
	resp, err := c.httpClient.Post(c.baseURL+"/conversations", "application/json", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create conversation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &conversation.Conversation{
		ID:       result["id"],
		Messages: make([]*message.Message, 0),
	}, nil
}

func (c *Client) ListConversations() ([]conversation.ConversationMetadata, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/conversations")
	if err != nil {
		return nil, fmt.Errorf("failed to list conversations: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var conversations []conversation.ConversationMetadata
	if err := json.NewDecoder(resp.Body).Decode(&conversations); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return conversations, nil
}

func (c *Client) GetConversation(id string) (*conversation.Conversation, error) {
	resp, err := c.httpClient.Get(c.baseURL + "/conversations/" + id)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, conversation.ErrConversationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error: %s", string(body))
	}

	var conv conversation.Conversation
	if err := json.NewDecoder(resp.Body).Decode(&conv); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &conv, nil
}

func (c *Client) SaveConversation(conv *conversation.Conversation) error {
	jsonData, err := json.Marshal(conv)
	if err != nil {
		return fmt.Errorf("failed to marshal conversation: %w", err)
	}

	url := fmt.Sprintf("%s/conversations/%s", c.baseURL, conv.ID)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to save conversation: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return conversation.ErrConversationNotFound
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server error: %s", string(body))
	}

	return nil
}

func (c *Client) GetLatestConversationID() (string, error) {
	conversations, err := c.ListConversations()
	if err != nil {
		return "", err
	}

	if len(conversations) == 0 {
		return "", conversation.ErrConversationNotFound
	}

	return conversations[0].ID, nil
}
