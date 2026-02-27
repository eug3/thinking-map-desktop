/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/cmd/app.go
 */
package cmd

import (
	"context"
	"log/slog"

	"github.com/PGshen/thinking-map-desktop/bindings"
	"github.com/PGshen/thinking-map-desktop/database"
	"github.com/PGshen/thinking-map-desktop/events"
	"github.com/PGshen/thinking-map-desktop/models"
	"github.com/PGshen/thinking-map-desktop/repositories"
	"github.com/PGshen/thinking-map-desktop/services"
	"github.com/wailsapp/wails/v3/pkg/application"
)

// AppContext represents the main application context
// It implements the ServiceStartup interface for Wails v3
type AppContext struct {
	// Core components
	DB     *database.DatabaseManager
	Events *events.EventManager

	// Repositories
	MapRepo     repositories.ThinkingMapRepository
	NodeRepo    repositories.ThinkingNodeRepository
	MessageRepo repositories.MessageRepository

	// Services
	ConfigService *services.ConfigService

	// Bindings (exposed as services)
	MapService      *bindings.MapBinding
	NodeService     *bindings.NodeBinding
	ThinkingService *bindings.ThinkingBinding
	ConfigBinding   *bindings.ConfigBinding
}

// NewAppContext creates a new app context with empty binding instances
// The bindings will be initialized during ServiceStartup
func NewAppContext() *AppContext {
	return &AppContext{
		MapService:      &bindings.MapBinding{},
		NodeService:     &bindings.NodeBinding{},
		ThinkingService: &bindings.ThinkingBinding{},
		ConfigBinding:   &bindings.ConfigBinding{},
	}
}

// Maps returns the map service binding
func (a *AppContext) Maps() *bindings.MapBinding {
	return a.MapService
}

// Nodes returns the node service binding
func (a *AppContext) Nodes() *bindings.NodeBinding {
	return a.NodeService
}

// Thinking returns the thinking service binding
func (a *AppContext) Thinking() *bindings.ThinkingBinding {
	return a.ThinkingService
}

// GetEvents returns the event manager (for setup from main.go)
func (a *AppContext) GetEvents() *events.EventManager {
	return a.Events
}

// ServiceStartup is called by Wails v3 during application startup
func (a *AppContext) ServiceStartup(ctx context.Context, options application.ServiceOptions) error {
	// Initialize SQLite database
	dbPath := "./data/thinking_map.db"
	db, err := database.NewDatabaseManager(dbPath)
	if err != nil {
		return err
	}
	a.DB = db

	// Run migrations
	if err := db.Migrate(
		&models.ThinkingMap{},
		&models.ThinkingNode{},
		&models.Message{},
	); err != nil {
		return err
	}
	slog.Info("Database migrated successfully", "path", dbPath)

	// Initialize event manager
	a.Events = events.NewEventManager()

	// Initialize repositories
	a.MapRepo = repositories.NewThinkingMapRepository(db.DB())
	a.NodeRepo = repositories.NewThinkingNodeRepository(db.DB())
	a.MessageRepo = repositories.NewMessageRepository(db.DB())

	// Initialize config service
	a.ConfigService = services.NewConfigService("./data")

	// Initialize LLM client with saved config
	llmCfg := a.ConfigService.GetLLMConfig()
	llmClient := services.NewLLMClient(&llmCfg)

	// Initialize services
	mapService := services.NewMapService(a.MapRepo)
	nodeService := services.NewNodeService(a.NodeRepo, a.MapRepo)
	thinkingService := services.NewThinkingService(
		a.NodeRepo,
		a.MapRepo,
		a.MessageRepo,
		a.Events,
		llmClient,
	)

	// Initialize bindings (exposed to frontend)
	bindings.InitMapBinding(a.MapService, mapService)
	bindings.InitNodeBinding(a.NodeService, nodeService)
	bindings.InitThinkingBinding(a.ThinkingService, thinkingService, a.Events)
	bindings.InitConfigBinding(a.ConfigBinding, a.ConfigService, func(cfg services.LLMConfig) {
		llmClient.UpdateConfig(&cfg)
		slog.Info("LLM config updated", "model", cfg.Model)
	})

	slog.Info("Application context initialized")
	return nil
}

// ServiceShutdown is called by Wails v3 during application shutdown
func (a *AppContext) ServiceShutdown() error {
	if a.DB != nil {
		a.DB.Close()
	}
	slog.Info("Application context stopped")
	return nil
}
