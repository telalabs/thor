package engine

import (
    "context"
    "fmt"

    "github.com/telalabs/thor/id"
    "github.com/telalabs/thor/llm"
    "github.com/telalabs/thor/logger"
    "github.com/telalabs/thor/manager"
    "github.com/telalabs/thor/options"
    "github.com/telalabs/thor/stores"

    "gorm.io/gorm"
)

// ValidateRequiredFields checks if all required fields in the Engine are set.
func (e *Engine) ValidateRequiredFields() error {
    if e.ctx == nil {
        return fmt.Errorf("context is required")
    }
    if e.db == nil {
        return fmt.Errorf("database connection is required")
    }
    if e.logger == nil {
        return fmt.Errorf("logger is required")
    }
    if e.actorStore == nil {
        return fmt.Errorf("actor store is required")
    }
    if e.sessionStore == nil {
        return fmt.Errorf("session store is required")
    }
    if e.interactionFragmentStore == nil {
        return fmt.Errorf("interaction fragment store is required")
    }
    if e.ID == "" {
        return fmt.Errorf("ID is required")
    }
    if e.Name == "" {
        return fmt.Errorf("name is required")
    }
    if e.llmClient == nil {
        return fmt.Errorf("LLM client is required")
    }
    return nil
}

// WithContext sets the context for the Engine.
func WithContext(ctx context.Context) options.Option[Engine] {
    return func(e *Engine) error {
        e.ctx = ctx
        return nil
    }
}

// WithDB sets the database connection for the Engine.
func WithDB(db *gorm.DB) options.Option[Engine] {
    return func(e *Engine) error {
        e.db = db
        return nil
    }
}

// WithLogger sets the logger for the Engine.
func WithLogger(logger *logger.Logger) options.Option[Engine] {
    return func(e *Engine) error {
        e.logger = logger
        return nil
    }
}

// WithIdentifier sets the ID and name for the Engine.
func WithIdentifier(id id.ID, name string) options.Option[Engine] {
    return func(e *Engine) error {
        e.ID = id
        e.Name = name
        return nil
    }
}

// WithInteractionFragmentStore sets the interaction fragment store for the Engine.
func WithInteractionFragmentStore(store *stores.FragmentStore) options.Option[Engine] {
    return func(e *Engine) error {
        e.interactionFragmentStore = store
        return nil
    }
}

// WithActorStore sets the actor store for the Engine.
func WithActorStore(store *stores.ActorStore) options.Option[Engine] {
    return func(e *Engine) error {
        e.actorStore = store
        return nil
    }
}

// WithSessionStore sets the session store for the Engine.
func WithSessionStore(store *stores.SessionStore) options.Option[Engine] {
    return func(e *Engine) error {
        e.sessionStore = store
        return nil
    }
}

// WithManagers sets the list of managers for the Engine, checking for duplicates and dependencies.
func WithManagers(_managers ...manager.Manager) options.Option[Engine] {
    return func(e *Engine) error {
        available := make(map[manager.ManagerID]manager.Manager)
        for _, m := range _managers {
            id := m.GetID()
            if _, exists := available[id]; exists {
                return fmt.Errorf("duplicate manager with ID %s", id)
            }
            available[id] = m
        }

        for _, m := range _managers {
            for _, dep := range m.GetDependencies() {
                if _, ok := available[dep]; !ok {
                    return fmt.Errorf("manager %s requires manager %s which was not provided", m.GetID(), dep)
                }
            }
        }

        e.managers = _managers
        return nil
    }
}

// WithManagerOrder sets the execution order for managers in the Engine.
func WithManagerOrder(order []manager.ManagerID) options.Option[Engine] {
    return func(e *Engine) error {
        managerMap := make(map[manager.ManagerID]bool)
        for _, m := range e.managers {
            managerMap[m.GetID()] = true
        }

        for _, id := range order {
            if !managerMap[id] {
                return fmt.Errorf("manager %s specified in order but not provided", id)
            }
        }

        e.managerOrder = order
        return nil
    }
}

// WithLLMClient sets the LLM client for the Engine.
func WithLLMClient(client *llm.LLMClient) options.Option[Engine] {
    return func(e *Engine) error {
        e.llmClient = client
        return nil
    }
}
