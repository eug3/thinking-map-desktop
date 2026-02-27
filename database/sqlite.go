/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/database/sqlite.go
 */
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DatabaseManager manages the database connection
type DatabaseManager struct {
	db    *gorm.DB
	sqlDB *sql.DB
	path  string
}

// NewDatabaseManager creates a new database manager
func NewDatabaseManager(dbPath string) (*DatabaseManager, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite connection with WAL mode for better performance
	dsn := fmt.Sprintf("%s?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=foreign_keys(true)&_pragma=cache_size(-64000)&_pragma=temp_store(memory)", dbPath)

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)

	return &DatabaseManager{
		db:    db,
		sqlDB: sqlDB,
		path:  dbPath,
	}, nil
}

// DB returns the GORM DB instance
func (m *DatabaseManager) DB() *gorm.DB {
	return m.db
}

// SQL returns the raw SQL DB instance
func (m *DatabaseManager) SQL() *sql.DB {
	return m.sqlDB
}

// Migrate runs auto-migration for given models
func (m *DatabaseManager) Migrate(models ...interface{}) error {
	return m.db.AutoMigrate(models...)
}

// Close closes the database connection
func (m *DatabaseManager) Close() error {
	if m.sqlDB != nil {
		return m.sqlDB.Close()
	}
	return nil
}

// Path returns the database file path
func (m *DatabaseManager) Path() string {
	return m.path
}
