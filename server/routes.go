package server

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/honganh1206/clue/server/conversation"
	_ "github.com/mattn/go-sqlite3"
)

type Server struct {
	addr net.Addr
}

func initConversationDsn() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatal("Failed to get home directory:", err)
	}

	dsn := filepath.Join(homeDir, ".local", ".clue", "conversation.db")
	return dsn
}

func Serve(ln net.Listener) error {
	dsn := initConversationDsn()
	db, err := conversation.InitDB(dsn)
	if err != nil {
		log.Fatalf("Failed to initialize database: %s", err.Error())
	}
	defer db.Close()

	srv := &Server{
		addr: ln.Addr(),
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Register conversation handlers
	mux.HandleFunc("/conversations", srv.handleConversations)
	mux.HandleFunc("/conversations/", srv.handleConversationByID)

	server := &http.Server{Handler: mux, Addr: ":11435"}
	return server.Serve(ln)
}

func (s *Server) handleConversations(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		s.createConversation(w, r)
	case http.MethodGet:
		s.listConversations(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleConversationByID(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path
	path := strings.TrimPrefix(r.URL.Path, "/conversations/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 || parts[0] == "" {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	conversationID := parts[0]

	switch r.Method {
	case http.MethodGet:
		s.getConversation(w, r, conversationID)
	case http.MethodPut:
		s.saveConversation(w, r, conversationID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) createConversation(w http.ResponseWriter, r *http.Request) {
	dsn := initConversationDsn()
	db, err := conversation.InitDB(dsn)
	if err != nil {
		http.Error(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	models := NewModels(db)
	conv, err := NewConversation()
	if err != nil {
		http.Error(w, "Failed to create conversation", http.StatusInternalServerError)
		return
	}

	if err := models.Conversations.SaveTo(conv); err != nil {
		http.Error(w, "Failed to save conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"id": conv.ID})
}

func (s *Server) listConversations(w http.ResponseWriter, r *http.Request) {
	dsn := initConversationDsn()
	db, err := conversation.InitDB(dsn)
	if err != nil {
		http.Error(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	models := NewModels(db)
	conversations, err := models.Conversations.List()
	if err != nil {
		http.Error(w, "Failed to list conversations", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conversations)
}

func (s *Server) getConversation(w http.ResponseWriter, r *http.Request, id string) {
	dsn := initConversationDsn()
	db, err := conversation.InitDB(dsn)
	if err != nil {
		http.Error(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	models := NewModels(db)
	conv, err := models.Conversations.Load(id)
	if err != nil {
		if err == conversation.ErrConversationNotFound {
			http.Error(w, "Conversation not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to load conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(conv)
}

func (s *Server) saveConversation(w http.ResponseWriter, r *http.Request, conversationID string) {
	dsn := initConversationDsn()
	db, err := conversation.InitDB(dsn)
	if err != nil {
		http.Error(w, "Failed to initialize database", http.StatusInternalServerError)
		return
	}
	defer db.Close()

	models := NewModels(db)
	var conv conversation.Conversation
	if err := json.NewDecoder(r.Body).Decode(&conv); err != nil {
		http.Error(w, "Invalid conversation format", http.StatusBadRequest)
		return
	}

	// Ensure the conversation ID matches the URL parameter
	if conv.ID != conversationID {
		http.Error(w, "Conversation ID mismatch", http.StatusBadRequest)
		return
	}

	// Save the entire conversation
	if err := models.Conversations.SaveTo(&conv); err != nil {
		http.Error(w, "Failed to save conversation", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "conversation saved"})
}
