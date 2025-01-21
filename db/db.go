package db

import (
    "fmt"
    "log"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// NewDatabase initializes a new database connection with GORM using PostgreSQL.
// It also ensures the vector extension is enabled, checks its version, 
// auto-migrates schemas, and creates fragment tables.
func NewDatabase(url string) (*gorm.DB, error) {
    db, err := gorm.Open(postgres.Open(url), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Silent),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Enable vector extension if not exists
    if err := enableVectorExtension(db); err != nil {
        return nil, err
    }

    // Verify vector extension is properly installed
    if err := checkVectorExtensionVersion(db); err != nil {
        return nil, err
    }

    // Auto-migrate the schema for Actor and Session models
    if err := autoMigrateSchemas(db); err != nil {
        return nil, err
    }

    // Create fragment tables
    if err := CreateFragmentTables(db); err != nil {
        return nil, err
    }

    return db, nil
}

// enableVectorExtension checks if the vector extension exists, 
// and creates it if it does not.
func enableVectorExtension(db *gorm.DB) error {
    if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
        return fmt.Errorf("failed to enable vector extension: %w", err)
    }
    return nil
}

// checkVectorExtensionVersion verifies that the vector extension is installed correctly.
func checkVectorExtensionVersion(db *gorm.DB) error {
    var version string
    if err := db.Raw("SELECT extversion FROM pg_extension WHERE extname = 'vector'").Scan(&version).Error; err != nil {
        return fmt.Errorf("vector extension not properly installed: %w", err)
    }
    log.Printf("pgvector extension version: %s", version)
    return nil
}

// autoMigrateSchemas handles the migration of the schema for specified models.
func autoMigrateSchemas(db *gorm.DB) error {
    if err := db.AutoMigrate(&Actor{}, &Session{}); err != nil {
        return fmt.Errorf("failed to migrate schemas: %w", err)
    }
    return nil
}

// CreateFragmentTables creates tables for the fragments if they do not exist.
func CreateFragmentTables(db *gorm.DB) error {
    for _, table := range fragmentTables {
        if !db.Migrator().HasTable(string(table)) {
            if err := db.Migrator().CreateTable(&Fragment{}, "table_name", string(table)); err != nil {
                return fmt.Errorf("failed to create %s table: %w", table, err)
            }
        }
    }
    return nil
}
