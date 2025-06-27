package conversation

import (
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/honganh1206/clue/message"
	"github.com/honganh1206/clue/server/db"
	"github.com/honganh1206/clue/utils"
)

//go:embed schema.sql
var schemaSQL string

var (
	ErrConversationNotFound = errors.New("history: conversation not found")
)

type Conversation struct {
	ID        string
	Messages  []*message.Message
	CreatedAt time.Time
}

type ConversationMetadata struct {
	ID                string
	LatestMessageTime time.Time
	MessageCount      int
	CreatedAt         time.Time
}

type ConversationModel struct {
	DB *sql.DB
}

func InitDB(dsn string) (*sql.DB, error) {
	dbConfig := db.Config{
		Dsn:          dsn,
		MaxOpenConns: 25,
		MaxIdleConns: 25,
		MaxIdleTime:  "15m",
	}

	conversationDb, err := db.OpenDB(dbConfig, schemaSQL)
	if err != nil {
		return nil, err
	}

	return conversationDb, nil
}

func NewConversation() (*Conversation, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	return &Conversation{
		ID:        id.String(),
		Messages:  make([]*message.Message, 0),
		CreatedAt: time.Now(),
	}, nil
}

func (cm ConversationModel) Append(c *Conversation, msg message.Message) {
	now := time.Now()
	sequence := len(c.Messages)

	msg.CreatedAt = now
	msg.Sequence = sequence

	c.Messages = append(c.Messages, &msg)
}

func (cm ConversationModel) SaveTo(c *Conversation) error {
	// Begin a transaction
	tx, err := cm.DB.Begin()
	if err != nil {
		return err
	}

	// TODO: Do I need to init a context for timeouts/graceful cancellation/tracing and logging?

	query := `
	INSERT OR IGNORE INTO conversations (id, created_at)
	VALUES(?, ?);
	`

	if _, err = tx.Exec(query, c.ID, c.CreatedAt); err != nil {
		tx.Rollback()
		return err
	}

	// FIXME: Currently delete and re-insert all messages, extremely inefficient
	// There should be a lastSavedIndex to insert the latest message. Should it be a column?
	query = `
	DELETE FROM messages WHERE conversation_id = ?;
	`

	if _, err = tx.Exec(query, c.ID); err != nil {
		tx.Rollback()
		return err
	}

	query = `
	INSERT INTO messages (conversation_id, sequence_number, payload, created_at)
	VALUES (?, ?, ?, ?);
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for i, msg := range c.Messages {
		jsonBytes, jsonErr := json.Marshal(msg)
		if jsonErr != nil {
			tx.Rollback()
			return jsonErr
		}
		payloadString := string(jsonBytes)
		_, err = stmt.Exec(c.ID, i, payloadString, msg.CreatedAt)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (cm ConversationModel) List() ([]ConversationMetadata, error) {
	query := `
		SELECT
			c.id,
			c.created_at,
			COUNT(m.id) as message_count,
			COALESCE(MAX(m.created_at), c.created_at) as latest_message_at
		FROM
			conversations c
		LEFT JOIN
			messages m ON c.id = m.conversation_id
		GROUP BY
			c.id
		ORDER BY
			latest_message_at DESC;
	`

	rows, err := cm.DB.Query(query)
	if err != nil {
		// Check for missing tables
		var tableCheck string
		errTable := cm.DB.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name='conversations'").Scan(&tableCheck)
		if errTable == sql.ErrNoRows {
			return []ConversationMetadata{}, nil // No 'conversations' table, so no conversations
		}
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}

	defer rows.Close()

	var metadataList []ConversationMetadata
	for rows.Next() {
		var meta ConversationMetadata
		var createdAt string
		var latestTimestamp string

		if err := rows.Scan(&meta.ID, &createdAt, &meta.MessageCount, &latestTimestamp); err != nil {
			return nil, fmt.Errorf("failed to scan conversation metadata: %w", err)
		}
		meta.CreatedAt, err = utils.ParseTimeWithFallback(createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse conversation created_at: %w", err)
		}

		meta.LatestMessageTime, err = utils.ParseTimeWithFallback(latestTimestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to parse latest_message_timestamp: %w", err)
		}
		metadataList = append(metadataList, meta)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return metadataList, nil

}

func (cm ConversationModel) LatestID() (string, error) {
	query := `
		SELECT id FROM conversations ORDER BY created_at DESC LIMIT 1
	`

	var id string
	err := cm.DB.QueryRow(query).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrConversationNotFound // Return custom error
		}
		return "", fmt.Errorf("failed to query for latest conversation ID: %w", err)
	}

	return id, nil
}

func (cm ConversationModel) Load(id string) (*Conversation, error) {
	query := `
		SELECT created_at FROM conversations WHERE id = ?
	`

	conv := &Conversation{ID: id, Messages: make([]*message.Message, 0)}

	err := cm.DB.QueryRow(query, id).Scan(&conv.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrConversationNotFound
		}
		return nil, fmt.Errorf("failed to query conversation metadata for ID '%s': %w", id, err)
	}

	query = `
		SELECT
			payload
		FROM
			messages WHERE conversation_id = ?
		ORDER BY
			sequence_number ASC
	`

	rows, err := cm.DB.Query(query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to query messages for conversation ID '%s': %w", id, err)
	}
	defer rows.Close()

	var msgs []*message.Message

	for rows.Next() {
		var payload []byte

		if err := rows.Scan(&payload); err != nil {
			return nil, fmt.Errorf("failed to scan message for conversation ID '%s': %w", id, err)
		}

		if err := json.Unmarshal(payload, &msgs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal temp message payload for conversation ID '%s': %w", id, err)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during message rows iteration for conversation ID '%s': %w", id, err)
	}

	for _, msg := range msgs {
		conv.Messages = append(conv.Messages, msg)
	}

	return conv, nil

}
