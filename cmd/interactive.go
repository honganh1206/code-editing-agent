package cmd

import (
	"bufio"
	"context"
	"log"
	"os"

	"github.com/honganh1206/clue/agent"
	"github.com/honganh1206/clue/api"
	"github.com/honganh1206/clue/inference"
	"github.com/honganh1206/clue/prompts"
	"github.com/honganh1206/clue/server/conversation"
	"github.com/honganh1206/clue/tools"
)

func interactive(ctx context.Context, convID string, modelConfig inference.ModelConfig, client *api.Client) error {
	model, err := inference.Init(modelConfig)
	if err != nil {
		log.Fatalf("Failed to initialize model: %s", err.Error())
	}

	scanner := bufio.NewScanner(os.Stdin)
	getUserMsg := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	toolDefs := []tools.ToolDefinition{
		tools.ReadFileDefinition,
		tools.ListFilesDefinition,
		tools.EditFileDefinition,
	}

	var conv *conversation.Conversation

	if convID != "" {
		conv, err = client.GetConversation(convID)
		if err != nil {
			return err
		}
	} else {
		conv, err = client.CreateConversation()
		if err != nil {
			return err
		}
	}
	a := agent.New(model, getUserMsg, conv, toolDefs, prompts.ClaudeSystemPrompt(), client)

	// In production, use Background() as the final root context()
	// For dev env, TODO for temporary scaffolding
	err = a.Run(ctx)

	if err != nil {
		return err
	}

	return nil
}
