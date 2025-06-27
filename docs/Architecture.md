# Architecture

Multi-binary architecture like Ollama

CLI, server and model engine are separate

Root `main.go` as stateless command-line interface and client operations. The CLI starts conversations and manages user I/O

The CLI client `clue serve` send HTTP request to `server/`

`app/main.go` as server daemon and API service (background processing). The server performs CRUD operations on conversations and messages (also inference routing?)

## High-level Components

1. Terminal Interface Layer (`cmd/`)
   Handles raw input/output with the terminal
   Manages command history and editing
   Implements custom rendering for code, tables, and other structured outputs
   Captures and redirects system outputs from executed commands
2. Command Processing Engine (`inference/, agent/`)
   Parses natural language inputs
   Identifies command intents and parameters
   Routes requests to appropriate handlers
   Manages conversation context and history
3. Codebase Analysis System (?)
   Scans and indexes project files
   Builds dependency graphs and structure maps
   Performs text and semantic searching
   Monitors file system changes
4. Execution Environment (?)
   Executes shell commands securely
   Captures and parses command outputs
   Manages environment variables and context
   Handles background and long-running processes
5. AI Integration Layer (`server/, app/`)
   Formats requests to the Claude API
   Processes and parses AI responses
   Manages AI context and history
   Handles authentication and API communication
6. File Operation System (?) (`server/`)
   Reads and writes files with appropriate permissions
   Generates diffs and patches
   Implements version control operations
   Handles file watching and change detection

## Data Flow

1. Input Processing Flow

Terminal input → Command parser → Intent classification → Handler selection → Action execution

2. Context Gathering Flow

Command intent → Context requirements → File system queries → Codebase analysis → Context compilation

3. AI Request Flow

User intent + Context → Request formatting → API authentication → Request transmission → Response reception → Response parsing

4. Response Handling Flow

Parsed response → Action extraction → Command generation → Execution → Output capture → Formatted display

## Impementation Details

### Programming Language and Runtime

Built with Go

### Key Dependencies

Terminal rendering libraries
File system utilities

### Design Patterns (!!)

Command Pattern for action encapsulation (commands)
Observer Pattern for system monitoring (metrics, logging)
Factory Pattern for handler creation
Adapter Pattern for external integrations

## Execution Flow

1. Initialization Phase (`cmd.go`)

Environment validation
Configuration loading
Authentication verification
Workspace scanning

2. Main Execution Loop (`agent.go`)

Input capture
Command processing
Context gathering
AI request/response handling
Action execution
Result presentation

3. Termination Phase (`lifecycle.go`)

Session state saving
Resource cleanup
Telemetry submission (if enabled)

## Error Handling

Hierarchical error classification
Graceful degradation for non-critical failures
Comprehensive logging to JSON files
User-friendly error messages with suggestions
Automatic retry mechanisms where appropriate

## Extensibility

Custom command handlers
Project-specific configuration
Tool integration adapters
Language-specific processors
