package conversation

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func createTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tempDir, err := os.MkdirTemp("", "conversation_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}

	t.Cleanup(func() {
		os.RemoveAll(tempDir)
	})

	testDBPath := filepath.Join(tempDir, "test.db")

	t.Cleanup(func() {
	})

	db, err := InitDB(testDBPath)
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestConversation_Append(t *testing.T) {
	conv, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	msg := MessagePartRequest{
		Role: UserRole,
		Content: []ContentBlock{
			NewTextContentBlock("Hello, world!"),
		},
	}

	conv.Append(msg)

	if len(conv.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(conv.Messages))
	}

	appended := conv.Messages[0]
	if appended.Role != UserRole {
		t.Errorf("Expected role %s, got %s", UserRole, appended.Role)
	}
	if appended.Sequence != 0 {
		t.Errorf("Expected sequence 0, got %d", appended.Sequence)
	}
	if appended.CreatedAt.IsZero() {
		t.Error("CreatedAt was not set")
	}

	msg2 := MessagePartRequest{
		Role: AssistantRole,
		Content: []ContentBlock{
			NewTextContentBlock("Hello back!"),
		},
	}

	conv.Append(msg2)

	if len(conv.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(conv.Messages))
	}

	appended2 := conv.Messages[1]
	if appended2.Sequence != 1 {
		t.Errorf("Expected sequence 1, got %d", appended2.Sequence)
	}
	if appended2.CreatedAt.Before(appended.CreatedAt) {
		t.Error("Second message CreatedAt should be after first message")
	}
}

func TestConversation_SaveTo(t *testing.T) {
	db := createTestDB(t)

	conv, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	conv.Append(MessagePartRequest{
		Role: UserRole,
		Content: []ContentBlock{
			NewTextContentBlock("First message"),
		},
	})

	conv.Append(MessagePartRequest{
		Role: AssistantRole,
		Content: []ContentBlock{
			NewTextContentBlock("Second message"),
		},
	})

	if err := conv.SaveTo(db); err != nil {
		t.Fatalf("SaveTo() failed: %v", err)
	}

	var savedID string
	var savedCreatedAt time.Time
	err = db.QueryRow("SELECT id, created_at FROM conversations WHERE id = ?", conv.ID).
		Scan(&savedID, &savedCreatedAt)
	if err != nil {
		t.Fatalf("Failed to query saved conversation: %v", err)
	}

	if savedID != conv.ID {
		t.Errorf("Expected ID %s, got %s", conv.ID, savedID)
	}

	rows, err := db.Query("SELECT sequence_number, payload FROM messages WHERE conversation_id = ? ORDER BY sequence_number", conv.ID)
	if err != nil {
		t.Fatalf("Failed to query saved messages: %v", err)
	}
	defer rows.Close()

	messageCount := 0
	for rows.Next() {
		var sequence int
		var payload string
		if err := rows.Scan(&sequence, &payload); err != nil {
			t.Fatalf("Failed to scan message row: %v", err)
		}

		if sequence != messageCount {
			t.Errorf("Expected sequence %d, got %d", messageCount, sequence)
		}

		messageCount++
	}

	if messageCount != len(conv.Messages) {
		t.Errorf("Expected %d saved messages, got %d", len(conv.Messages), messageCount)
	}
}

func TestConversation_SaveTo_DuplicateConversation(t *testing.T) {
	db := createTestDB(t)

	conv, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	conv.Append(MessagePartRequest{
		Role: UserRole,
		Content: []ContentBlock{
			NewTextContentBlock("Test message"),
		},
	})

	// Save conversation first time
	if err := conv.SaveTo(db); err != nil {
		t.Fatalf("First SaveTo() failed: %v", err)
	}

	// Add another message and save again
	conv.Append(MessagePartRequest{
		Role: AssistantRole,
		Content: []ContentBlock{
			NewTextContentBlock("Response message"),
		},
	})

	if err := conv.SaveTo(db); err != nil {
		t.Fatalf("Second SaveTo() failed: %v", err)
	}

	// Verify only one conversation record exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM conversations WHERE id = ?", conv.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count conversations: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 conversation record, got %d", count)
	}

	// Verify correct number of messages
	err = db.QueryRow("SELECT COUNT(*) FROM messages WHERE conversation_id = ?", conv.ID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count messages: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected 2 message records, got %d", count)
	}
}

func TestList(t *testing.T) {
	db := createTestDB(t)

	// Test empty database
	metadataList, err := List(db)
	if err != nil {
		t.Fatalf("List() failed on empty database: %v", err)
	}
	if len(metadataList) != 0 {
		t.Errorf("Expected 0 conversations, got %d", len(metadataList))
	}

	// Create and save conversations
	conv1, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	conv1.Append(MessagePartRequest{
		Role: UserRole,
		Content: []ContentBlock{
			NewTextContentBlock("First conversation message"),
		},
	})

	conv2, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	conv2.Append(MessagePartRequest{
		Role: UserRole,
		Content: []ContentBlock{
			NewTextContentBlock("Second conversation message"),
		},
	})
	conv2.Append(MessagePartRequest{
		Role: AssistantRole,
		Content: []ContentBlock{
			NewTextContentBlock("Response to second conversation"),
		},
	})

	// Save conversations
	if err := conv1.SaveTo(db); err != nil {
		t.Fatalf("SaveTo() failed for conv1: %v", err)
	}

	// Add a small delay to ensure different timestamps
	time.Sleep(1 * time.Millisecond)

	if err := conv2.SaveTo(db); err != nil {
		t.Fatalf("SaveTo() failed for conv2: %v", err)
	}

	// Test List function
	metadataList, err = List(db)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(metadataList) != 2 {
		t.Errorf("Expected 2 conversations, got %d", len(metadataList))
	}

	// Verify the conversations are ordered by latest message time (DESC)
	if metadataList[0].ID != conv2.ID {
		t.Errorf("Expected first conversation to be %s, got %s", conv2.ID, metadataList[0].ID)
	}
	if metadataList[1].ID != conv1.ID {
		t.Errorf("Expected second conversation to be %s, got %s", conv1.ID, metadataList[1].ID)
	}

	// Verify message counts
	if metadataList[0].MessageCount != 2 {
		t.Errorf("Expected conv2 to have 2 messages, got %d", metadataList[0].MessageCount)
	}
	if metadataList[1].MessageCount != 1 {
		t.Errorf("Expected conv1 to have 1 message, got %d", metadataList[1].MessageCount)
	}

	// Verify timestamps are not zero
	for i, meta := range metadataList {
		if meta.CreatedAt.IsZero() {
			t.Errorf("Conversation %d CreatedAt is zero", i)
		}
		if meta.LatestMessageTime.IsZero() {
			t.Errorf("Conversation %d LatestMessageTime is zero", i)
		}
	}
}

func TestList_EmptyConversation(t *testing.T) {
	db := createTestDB(t)

	// Create conversation without messages
	conv, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Save empty conversation directly to database
	_, err = db.Exec("INSERT INTO conversations (id, created_at) VALUES (?, ?)", conv.ID, conv.CreatedAt)
	if err != nil {
		t.Fatalf("Failed to insert empty conversation: %v", err)
	}

	metadataList, err := List(db)
	if err != nil {
		t.Fatalf("List() failed: %v", err)
	}

	if len(metadataList) != 1 {
		t.Errorf("Expected 1 conversation, got %d", len(metadataList))
	}

	meta := metadataList[0]
	if meta.MessageCount != 0 {
		t.Errorf("Expected 0 messages, got %d", meta.MessageCount)
	}

	// For empty conversation, LatestMessageTime should equal CreatedAt
	if !meta.LatestMessageTime.Equal(meta.CreatedAt) {
		t.Errorf("Expected LatestMessageTime to equal CreatedAt for empty conversation")
	}
}

func TestLatestID(t *testing.T) {
	db := createTestDB(t)

	// Test empty database
	_, err := LatestID(db)
	if err != ErrConversationNotFound {
		t.Errorf("Expected ErrConversationNotFound, got %v", err)
	}

	// Create conversations with different creation times
	conv1, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Manually set creation time to ensure ordering
	conv1.CreatedAt = time.Now().Add(-1 * time.Hour)

	conv2, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	conv2.CreatedAt = time.Now()

	// Save conversations
	if err := conv1.SaveTo(db); err != nil {
		t.Fatalf("SaveTo() failed for conv1: %v", err)
	}
	if err := conv2.SaveTo(db); err != nil {
		t.Fatalf("SaveTo() failed for conv2: %v", err)
	}

	// Test LatestID function
	latestID, err := LatestID(db)
	if err != nil {
		t.Fatalf("LatestID() failed: %v", err)
	}

	if latestID != conv2.ID {
		t.Errorf("Expected latest ID to be %s, got %s", conv2.ID, latestID)
	}
}

func TestLoad(t *testing.T) {
	db := createTestDB(t)

	// Test loading non-existent conversation
	_, err := Load("non-existent-id", db)
	if err != ErrConversationNotFound {
		t.Errorf("Expected ErrConversationNotFound, got %v", err)
	}

	// Create and save a conversation with multiple message types
	conv, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Add text message
	conv.Append(MessagePartRequest{
		Role: UserRole,
		Content: []ContentBlock{
			NewTextContentBlock("Hello, this is a test message"),
		},
	})

	// Add tool use message
	toolInput := []byte(`{"query": "test"}`)
	conv.Append(MessagePartRequest{
		Role: AssistantRole,
		Content: []ContentBlock{
			NewToolUseContentBlock("tool-123", "search", toolInput),
		},
	})

	// Add tool result message
	conv.Append(MessagePartRequest{
		Role: UserRole,
		Content: []ContentBlock{
			NewToolResultContentBlock("tool-123", "Search results here", false),
		},
	})

	// Save conversation
	if err := conv.SaveTo(db); err != nil {
		t.Fatalf("SaveTo() failed: %v", err)
	}

	// Load conversation
	loadedConv, err := Load(conv.ID, db)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify basic properties
	if loadedConv.ID != conv.ID {
		t.Errorf("Expected ID %s, got %s", conv.ID, loadedConv.ID)
	}
	if !loadedConv.CreatedAt.Equal(conv.CreatedAt) {
		t.Errorf("Expected CreatedAt %v, got %v", conv.CreatedAt, loadedConv.CreatedAt)
	}
	if len(loadedConv.Messages) != len(conv.Messages) {
		t.Errorf("Expected %d messages, got %d", len(conv.Messages), len(loadedConv.Messages))
	}

	// Verify messages are loaded correctly
	for i, originalMsg := range conv.Messages {
		loadedMsg := loadedConv.Messages[i]

		if loadedMsg.Role != originalMsg.Role {
			t.Errorf("Message %d: Expected role %s, got %s", i, originalMsg.Role, loadedMsg.Role)
		}
		if loadedMsg.Sequence != originalMsg.Sequence {
			t.Errorf("Message %d: Expected sequence %d, got %d", i, originalMsg.Sequence, loadedMsg.Sequence)
		}
		if len(loadedMsg.Content) != len(originalMsg.Content) {
			t.Errorf("Message %d: Expected %d content blocks, got %d", i, len(originalMsg.Content), len(loadedMsg.Content))
		}

		// Verify content blocks
		for j, originalContent := range originalMsg.Content {
			loadedContent := loadedMsg.Content[j]

			switch original := originalContent.(type) {
			case TextContentBlock:
				loaded, ok := loadedContent.(TextContentBlock)
				if !ok {
					t.Errorf("Message %d, Content %d: Expected TextContentBlock, got %T", i, j, loadedContent)
					continue
				}
				if loaded.Text != original.Text {
					t.Errorf("Message %d, Content %d: Expected text %s, got %s", i, j, original.Text, loaded.Text)
				}

			case ToolUseContentBlock:
				loaded, ok := loadedContent.(ToolUseContentBlock)
				if !ok {
					t.Errorf("Message %d, Content %d: Expected ToolUseContentBlock, got %T", i, j, loadedContent)
					continue
				}
				if loaded.ID != original.ID {
					t.Errorf("Message %d, Content %d: Expected tool ID %s, got %s", i, j, original.ID, loaded.ID)
				}
				if loaded.Name != original.Name {
					t.Errorf("Message %d, Content %d: Expected tool name %s, got %s", i, j, original.Name, loaded.Name)
				}

			case ToolResultContentBlock:
				loaded, ok := loadedContent.(ToolResultContentBlock)
				if !ok {
					t.Errorf("Message %d, Content %d: Expected ToolResultContentBlock, got %T", i, j, loadedContent)
					continue
				}
				if loaded.ToolUseID != original.ToolUseID {
					t.Errorf("Message %d, Content %d: Expected tool use ID %s, got %s", i, j, original.ToolUseID, loaded.ToolUseID)
				}
				if loaded.IsError != original.IsError {
					t.Errorf("Message %d, Content %d: Expected is_error %v, got %v", i, j, original.IsError, loaded.IsError)
				}
			}
		}
	}
}

func TestLoad_EmptyConversation(t *testing.T) {
	db := createTestDB(t)

	// Create conversation without messages
	conv, err := New()
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Save empty conversation
	if err := conv.SaveTo(db); err != nil {
		t.Fatalf("SaveTo() failed: %v", err)
	}

	// Load conversation
	loadedConv, err := Load(conv.ID, db)
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	if loadedConv.ID != conv.ID {
		t.Errorf("Expected ID %s, got %s", conv.ID, loadedConv.ID)
	}
	if len(loadedConv.Messages) != 0 {
		t.Errorf("Expected 0 messages, got %d", len(loadedConv.Messages))
	}
}
