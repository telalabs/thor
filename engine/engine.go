package engine

import (
    "fmt"
    "time"

    "github.com/telalabs/thor/db"
    "github.com/telalabs/thor/id"
    "github.com/telalabs/thor/llm"
    "github.com/telalabs/thor/manager"
    "github.com/telalabs/thor/options"
    "github.com/telalabs/thor/state"
    toolkit "github.com/telalabs/toolkit/go"
    "github.com/pgvector/pgvector-go"
    "golang.org/x/sync/errgroup"
)

// New creates a new Engine instance with the provided options.
// Returns an error if required fields are missing or if actor creation fails.
func New(opts ...options.Option[Engine]) (*Engine, error) {
    e := &Engine{}
    if err := options.ApplyOptions(e, opts...); err != nil {
        return nil, fmt.Errorf("failed to create engine: %w", err)
    }

    if err := e.upsertActor(e.ID, e.Name, true); err != nil {
        return nil, fmt.Errorf("failed to upsert actor: %w", err)
    }

    return e, nil
}

// Process handles the processing of a new input through the runtime pipeline:
// 1. Retrieves actor and session information
// 2. Creates a copy of the input fragment
// 3. Executes all managers in parallel
// 4. Stores the processed input
// Returns an error if any step fails.
func (e *Engine) Process(currentState *state.State) error {
    input := currentState.Input

    e.logger.WithFields(map[string]interface{}{
        "input": input.ID,
    }).Info("Processing input")

    actor, err := e.actorStore.GetByID(input.ActorID)
    if err != nil {
        return fmt.Errorf("failed to get actor: %w", err)
    }

    session, err := e.sessionStore.GetByID(input.SessionID)
    if err != nil {
        return fmt.Errorf("failed to get session: %w", err)
    }

    inputCopy := e.createFragmentCopy(input, actor, session)

    currentState.Input = inputCopy

    errGroup := new(errgroup.Group)
    for _, m := range e.managers {
        m := m // Capture the loop variable
        errGroup.Go(func() error {
            return m.Process(currentState)
        })
    }

    if err := errGroup.Wait(); err != nil {
        return fmt.Errorf("failed to execute manager analysis: %w", err)
    }

    if err := e.interactionFragmentStore.Upsert(inputCopy); err != nil {
        return fmt.Errorf("failed to store input: %w", err)
    }

    return nil
}

// PostProcess handles the post-processing of a response:
// 1. Retrieves actor and session information
// 2. Creates a copy of the response fragment
// 3. Executes all managers in sequence
// 4. Stores the processed response
// Returns an error if any step fails.
func (e *Engine) PostProcess(response *db.Fragment, currentState *state.State) error {
    actor, err := e.actorStore.GetByID(response.ActorID)
    if err != nil {
        return fmt.Errorf("failed to get actor: %w", err)
    }

    session, err := e.sessionStore.GetByID(response.SessionID)
    if err != nil {
        return fmt.Errorf("failed to get session: %w", err)
    }

    responseCopy := e.createFragmentCopy(response, actor, session)

    currentState.Output = responseCopy

    if err := e.executeManagersInOrder(currentState, func(m manager.Manager) error {
        return m.PostProcess(currentState)
    }); err != nil {
        return fmt.Errorf("failed to execute manager actions: %w", err)
    }

    if err := e.interactionFragmentStore.Upsert(response); err != nil {
        return fmt.Errorf("failed to store response: %w", err)
    }

    return nil
}

// GenerateResponse creates a new response using the LLM:
// 1. Generates completion from provided messages
// 2. Creates embedding for the response
// 3. Builds response fragment with metadata
// Returns the response fragment and any error encountered.
func (e *Engine) GenerateResponse(messages []llm.Message, sessionID id.ID, tools ...toolkit.Tool) (*db.Fragment, error) {
    response, err := e.llmClient.GenerateCompletion(llm.CompletionRequest{
        Messages:    messages,
        ModelType:   llm.ModelTypeDefault,
        Temperature: 0.7,
        Tools:       tools,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to generate completion: %v", err)
    }

    embedding, err := e.llmClient.EmbedText(response.Content)
    if err != nil {
        return nil, fmt.Errorf("failed to create embedding for response: %v", err)
    }

    return &db.Fragment{
        ID:        id.New(),
        ActorID:   e.ID,
        SessionID: sessionID,
        Content:   response.Content,
        Embedding: pgvector.NewVector(embedding),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Metadata:  nil,
    }, nil
}

// StartBackgroundProcesses initiates background processes for all managers.
// Each manager's background process runs in its own goroutine.
func (e *Engine) StartBackgroundProcesses() {
    for _, m := range e.managers {
        go m.StartBackgroundProcesses()
    }
}

// StopBackgroundProcesses terminates background processes for all managers.
func (e *Engine) StopBackgroundProcesses() {
    for _, m := range e.managers {
        m.StopBackgroundProcesses()
    }
}

// AddManager adds a new manager to the runtime.
// Validates that:
// 1. The manager ID is not duplicate
// 2. All manager dependencies are available
// Returns an error if validation fails.
func (e *Engine) AddManager(newManager manager.Manager) error {
    for _, m := range e.managers {
        if m.GetID() == newManager.GetID() {
            return fmt.Errorf("duplicate manager with ID %s", newManager.GetID())
        }
    }

    available := make(map[manager.ManagerID]bool)
    for _, m := range e.managers {
        available[m.GetID()] = true
    }

    for _, dep := range newManager.GetDependencies() {
        if !available[dep] {
            return fmt.Errorf("manager %s requires manager %s which was not provided", newManager.GetID(), dep)
        }
    }

    e.managers = append(e.managers, newManager)
    return nil
}

// createFragmentCopy creates a copy of a fragment with provided actor and session data.
func (e *Engine) createFragmentCopy(fragment *db.Fragment, actor *db.Actor, session *db.Session) *db.Fragment {
    return &db.Fragment{
        ID:        fragment.ID,
        ActorID:   fragment.ActorID,
        SessionID: fragment.SessionID,
        Content:   fragment.Content,
        Metadata:  fragment.Metadata,
        Embedding: fragment.Embedding,
        Actor:     actor,
        Session:   session,
        CreatedAt: fragment.CreatedAt,
        UpdatedAt: fragment.UpdatedAt,
        DeletedAt: fragment.DeletedAt,
    }
}

// executeManagersInOrder runs managers in a specified order:
// 1. Creates a map for quick manager lookup
// 2. Uses managerOrder if specified, otherwise uses registration order
// 3. Executes each manager with the provided function
// Returns an error if any manager execution fails.
func (e *Engine) executeManagersInOrder(currentState *state.State, executeFn func(manager.Manager) error) error {
    managerMap := make(map[manager.ManagerID]manager.Manager)
    for _, m := range e.managers {
        managerMap[m.GetID()] = m
    }

    executionOrder := e.managerOrder
    if len(executionOrder) == 0 {
        executionOrder = make([]manager.ManagerID, len(e.managers))
        for i, m := range e.managers {
            executionOrder[i] = m.GetID()
        }
    }

    for _, managerID := range executionOrder {
        if manager, exists := managerMap[managerID]; exists {
            if err := executeFn(manager); err != nil {
                return fmt.Errorf("manager %s failed: %w", managerID, err)
            }
        }
    }

    return nil
}
