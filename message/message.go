package message

import (
	"encoding/json"
	"time"
)

type Message struct {
	Role string `json:"role"`
	// FIXME: Cannot unmarshal interface as not concrete type
	Content []ContentBlockUnion `json:"content"`
	// Optional as metadata
	ID        string    `json:"id,omitempty" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Sequence  int       `json:"sequence,omitempty" db:"sequence_number"`
}

const (
	UserRole      = "user"
	AssistantRole = "assistant"
)

const (
	TextType       = "text"
	ToolUseType    = "tool_use"
	ToolResultType = "tool_result"
)

type TextContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func NewTextContentBlock(text string) ContentBlockUnion {
	return ContentBlockUnion{
		Type: TextType,
		OfTextBlock: &TextContentBlock{
			Text: text,
		}}
}

type ToolUseContentBlock struct {
	Type     string          `json:"type"`
	Text     string          `json:"text"`
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	IsError  bool            `json:"is_error"`
	ToolCall bool            `json:"tool_call"`
}

func NewToolUseContentBlock(id, name string, input json.RawMessage) ContentBlockUnion {
	return ContentBlockUnion{
		Type: ToolUseType,
		OfToolUseBlock: &ToolUseContentBlock{
			ID:    id,
			Name:  name,
			Input: input,
		}}
}

type ToolResultContentBlock struct {
	Type      string `json:"type"`
	ToolUseID string `json:"tool_use_id"`
	Content   any    `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

func NewToolResultContentBlock(toolUseID string, content any, isError bool) ContentBlockUnion {
	return ContentBlockUnion{
		Type: ToolResultType,
		OfToolResultBlock: &ToolResultContentBlock{
			ToolUseID: toolUseID,
			Content:   content,
			IsError:   isError,
		}}

}

type ContentBlockUnion struct {
	Type              string                  `json:"type"`
	OfTextBlock       *TextContentBlock       `json:",omitzero,inline"`
	OfToolUseBlock    *ToolUseContentBlock    `json:",omitzero,inline"`
	OfToolResultBlock *ToolResultContentBlock `json:",omitzero,inline"`
}
