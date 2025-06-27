package server

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/honganh1206/clue/message"
	"github.com/honganh1206/clue/server/conversation"
)

type Models struct {
	Conversations *conversation.ConversationModel
}

func NewModels(db *sql.DB) *Models {
	return &Models{
		Conversations: &conversation.ConversationModel{DB: db},
	}
}

func NewConversation() (*conversation.Conversation, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &conversation.Conversation{
		ID:        id.String(),
		Messages:  make([]*message.Message, 0),
		CreatedAt: time.Now(),
	}, nil
}
