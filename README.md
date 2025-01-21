# Thor - Advanced LLM Framework

## Table of Contents
- [Overview](#Overview)
- [Core Features](#core-features)
- [Extension Points](#extension-points)
- [Quick Start](#quick-start)
- [Using Thor as a Module](#using-zen-as-a-module)

## Overview
Built in GO, Thor is a highly modular AI conversation engine that prioritizes platform independence and pluggable design. It offers an adaptable framework for creating conversational systems by:

- An architecture based on plugins that has hot-swappable parts
- Support for several suppliers of LLM (OpenAI, custom providers)
- Conversation management across platforms and an extensible manager system with unique behaviors
- Semantic storage using vectors and pgvector

## Core Features

### Plugin Architecture
- **Manager System**: Extend functionality through custom managers
  - Insight Manager: Extracts and maintains conversation insights
  - Personality Manager: Handles response behavior and style
  - Custom Managers: Add your own specialized behaviors

### State Management
- **Shared State System**: Centralized state management across components
  - Manager-specific data storage
  - Custom data injection
  - Cross-manager communication

### LLM Integration
- **Provider Abstraction**: Support for multiple LLM providers
  - Built-in OpenAI support
  - Extensible provider interface for custom LLMs
  - Configurable model selection per operation
  - Automatic fallback and retry handling

### Platform Support
- **Platform Agnostic Core**: 
  - Abstract conversation engine independent of platforms
  - Built-in support for CLI chat and Twitter
  - Extensible platform manager interface
  - Example implementations for new platform integration

### Storage Layer
- **Flexible Data Storage**:
  - PostgreSQL with pgvector for semantic search
  - GORM-based data models
  - Customizable fragment storage
  - Vector embedding support

### Toolkit/Function System
- **Pluggable Tool/Function Integration**:
  - Support for custom tool implementations
  - Built-in toolkit management
  - Function calling capabilities
  - Automatic tool response handling
  - State-aware tool execution

## Extension Points
1. **LLM Providers**: Add new AI providers by implementing the LLM interface
```go
type Provider interface {
    GenerateCompletion(context.Context, CompletionRequest) (string, error)
    GenerateJSON(context.Context, JSONRequest, interface{}) error
    EmbedText(context.Context, string) ([]float32, error)
}
```

2. **Managers**: Create new behaviors by implementing the Manager interface
```go
type Manager interface {
    GetID() ManagerID
    GetDependencies() []ManagerID
    Process(state *state.State) error
    PostProcess(state *state.State) error
    Context(state *state.State) ([]state.StateData, error)
    Store(fragment *db.Fragment) error
    StartBackgroundProcesses()
    StopBackgroundProcesses()
    RegisterEventHandler(callback EventCallbackFunc)
    triggerEvent(eventData EventData)
}
```

## Quick Start
1. Clone the repository
```bash 
git clone https://github.com/telalabs/thor
```   
2. Copy `.env.example` to `.env` and configure your environment variables
3. Install dependencies:
```bash
go mod download
```
4. Run the chat example:
```bash
go run examples/chat/main.go
```
6. Run the Twitter bot:
```bash
go run examples/twitter/main.go
```

## Environment Variables
```env
DB_URL=postgresql://user:password@localhost:5432/thor
OPENAI_API_KEY=your_openai_api_key

Platform-specific credentials as needed
```

## Architecture
The project follows a clean, modular architecture:

- `engine`: Core conversation engine
- `manager`: Plugin manager system
- `managers/*`: Built-in manager implementations
- `state`: Shared state management
- `llm`: LLM provider interfaces
- `stores`: Data storage implementations
- `tools/*`: Built-in tool implementations
- `examples/`: Reference implementations

## Using Thor as a Module

1. Add Thor to your Go project:
```bash
go get github.com/telalabs/thor
```

2. Import Thor in your code:
```go
import (
  "github.com/soralabs/thor/engine"
  "github.com/soralabs/thor/llm"
  "github.com/soralabs/thor/manager"
  "github.com/soralabs/thor/managers/personality"
  "github.com/soralabs/thor/managers/insight"
  ... etc
)
```

3. Basic usage example:
```go
// Initialize LLM client
llmClient, err := llm.NewLLMClient(llm.Config{
  ProviderType: llm.ProviderOpenAI,
  APIKey: os.Getenv("OPENAI_API_KEY"),
  ModelConfig: map[llm.ModelType]string{
    llm.ModelTypeDefault: openai.GPT4,
  },
  Logger: logger,
  Context: ctx,
})

// Create engine instance
engine, err := engine.New(
  engine.WithContext(ctx),
  engine.WithLogger(logger),
  engine.WithDB(db),
  engine.WithLLM(llmClient),
)

// Process input
state, err := engine.NewState(actorID, sessionID, "Your input text here")
if err != nil {
  log.Fatal(err)
}

response, err := engine.Process(state)
if err != nil {
  log.Fatal(err)
}
```

4. Available packages:
- `thor/engine`: Core conversation engine
- `thor/llm`: LLM provider interfaces and implementations
- `thor/manager`: Base manager system
- `thor/managers/*`: Built-in manager implementations
- `thor/state`: State management utilities
- `thor/stores`: Data storage implementations

For detailed examples, see the `examples/` directory in the repository.
