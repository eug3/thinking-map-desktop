/*
 * @Date: 2025-02-27
 * @FilePath: /thinking-map-desktop/main.go
 */
package main

import (
	"embed"
	"log"

	"github.com/PGshen/thinking-map-desktop/cmd"
	"github.com/PGshen/thinking-map-desktop/events"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	appCtx := cmd.NewAppContext()

	// Create Wails application
	wailsApp := application.New(application.Options{
		Name:        "ThinkingMap",
		Description: "AI-Powered Thinking Map Desktop Application",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Services: []application.Service{
			application.NewService(appCtx),
			application.NewService(appCtx.MapService),
			application.NewService(appCtx.NodeService),
			application.NewService(appCtx.ThinkingService),
			application.NewService(appCtx.ConfigBinding),
		},
	})

	// Create window
	wailsApp.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "ThinkingMap",
		Width:            1280,
		Height:           800,
		BackgroundColour: application.NewRGB(255, 255, 255),
		URL:              "/",
	})

	// Wire up global event emitter so Go events reach the frontend
	events.SetGlobalEmitFn(func(name string, data ...interface{}) {
		wailsApp.Event.Emit(name, data...)
	})

	// Run the application
	err := wailsApp.Run()
	if err != nil {
		log.Fatal(err)
	}
}
