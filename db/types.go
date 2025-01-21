package db

import (
    "encoding/json"
    "time"

    "github.com/soralabs/zen/id"
    "github.com/pgvector/pgvector-go"
    "gorm.io/gorm"

    "database/sql/driver"
    "errors"
)

// FragmentTable defines the different types of fragment tables in the database.
type FragmentTable string

const (
    FragmentTableInteraction FragmentTable = "interaction"
    FragmentTablePersonality FragmentTable = "personality"
    FragmentTableInsight     FragmentTable = "insight"
    FragmentTableTwitter     FragmentTable = "twitter"
)

var fragmentTables = []FragmentTable{
    FragmentTableInteraction,
    FragmentTablePersonality,
    FragmentTableInsight,
    FragmentTableTwitter,
}

// Metadata represents a JSON object stored in the database.
type Metadata map[string]interface{}

// Fragment represents a data fragment stored in one of the fragment tables.
type Fragment struct {
    ID        id.ID           `gorm:"type:uuid;primaryKey"`
    ActorID   id.ID           `gorm:"type:uuid;not null;index"`
    SessionID id.ID           `gorm:"type:uuid;not null;index"`
    Content   string          `gorm:"type:text;not null"`
    Metadata  Metadata        `gorm:"type:jsonb;not null;default:'{}'::jsonb"`
    Embedding pgvector.Vector `gorm:"type:vector(1536)"`

    Actor   *Actor   `gorm:"foreignKey:ActorID"`
    Session *Session `gorm:"foreignKey:SessionID"`

    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Actor represents an entity in the system with a unique ID and name.
type Actor struct {
    ID   id.ID  `gorm:"type:uuid;primaryKey"`
    Name string `gorm:"type:varchar(255);not null"`

    Assistant bool `gorm:"type:boolean;not null;default:false"`

    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Session represents a session with a unique ID.
type Session struct {
    ID id.ID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`

    CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}

// Value implements the driver.Valuer interface for Metadata.
func (m Metadata) Value() (driver.Value, error) {
    if m == nil {
        return json.Marshal(map[string]interface{}{})
    }
    return json.Marshal(m)
}

// Scan implements the sql.Scanner interface for Metadata.
func (m *Metadata) Scan(value interface{}) error {
    if value == nil {
        *m = make(map[string]interface{})
        return nil
    }

    bytes, ok := value.([]byte)
    if !ok {
        return errors.New("failed to unmarshal JSONB value: invalid type")
    }

    return json.Unmarshal(bytes, m)
}

// GetString retrieves a string value from Metadata, returning an empty string if not found or not a string.
func (m Metadata) GetString(key string) string {
    if val, ok := m[key].(string); ok {
        return val
    }
    return ""
}

// GetFloat retrieves a float64 value from Metadata, returning 0 if not found or not a float64.
func (m Metadata) GetFloat(key string) float64 {
    if val, ok := m[key].(float64); ok {
        return val
    }
    return 0
}

// GetBool retrieves a boolean value from Metadata, returning false if not found or not a bool.
func (m Metadata) GetBool(key string) bool {
    if val, ok := m[key].(bool); ok {
        return val
    }
    return false
}
