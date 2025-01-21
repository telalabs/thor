package engine

import (
    "fmt"

    "github.com/telalabs/thor/db"
    "github.com/telalabs/thor/id"
)

// UpsertSession creates or updates a session in the database.
// If the session ID already exists, it will be updated.
func (e *Engine) UpsertSession(sessionID id.ID) error {
    if err := e.sessionStore.Upsert(&db.Session{
        ID: sessionID,
    }); err != nil {
        return fmt.Errorf("failed to upsert session: %w", err)
    }
    return nil
}

// UpsertActor creates or updates an actor in the database.
// If the actor ID already exists, it will be updated with the new name and assistant status.
func (e *Engine) UpsertActor(actorID id.ID, actorName string, assistant bool) error {
    if err := e.actorStore.Upsert(&db.Actor{
        ID:        actorID,
        Name:      actorName,
        Assistant: assistant,
    }); err != nil {
        return fmt.Errorf("failed to upsert actor: %w", err)
    }
    return nil
}

// UpsertInteractionFragment creates or updates an interaction fragment in the database.
// If the fragment ID already exists, it will be updated with the new data.
func (e *Engine) UpsertInteractionFragment(fragment *db.Fragment) error {
    return e.interactionFragmentStore.Upsert(fragment)
}

// DoesInteractionFragmentExist checks if an interaction fragment exists in the database.
// Returns true if the fragment exists, false otherwise, along with any error encountered.
func (e *Engine) DoesInteractionFragmentExist(fragmentID id.ID) (bool, error) {
    fragment, err := e.interactionFragmentStore.GetByID(fragmentID)
    if err != nil {
        return false, fmt.Errorf("failed to check for fragment existence: %w", err)
    }
    // If fragment is nil, it means the fragment does not exist
    return fragment != nil, nil
}
