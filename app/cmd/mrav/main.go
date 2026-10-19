// Command mrav is the Move Viewer window.
package main

import (
	"context"
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	"app/config"
	"app/internal/core/move"
	"app/internal/core/move/repository/db/moverepo"
	"app/internal/core/move/service/movesvc"
	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/moveui"
	"app/internal/infrastructure/database"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/icons"
	"app/internal/port"
)

// options is the whole dependency graph. Each repository and service constructor returns its port, so
// none needs fx.Annotate; the fake source is the one thing provided as two types, because the demo panel
// needs its setters and the service needs only its port.
func options() fx.Option {
	return fx.Options(
		fx.Provide(provideConfig, database.ProvideDatabase),
		fx.Provide(moverepo.New, movesvc.New),
		fx.Provide(fakemonster.New, func(s *fakemonster.Source) port.MonsterSource { return s }),
		fx.Provide(newFyneApp, newScreen, newDemo, newWindow),
		fx.Invoke(refreshLoop),
	)
}

// main starts the graph by hand instead of fx.Run: Fyne must own the main thread, and fx.Run would block it.
func main() {
	var win fyne.Window
	app := fx.New(options(), fx.Populate(&win))
	if err := app.Err(); err != nil {
		log.Fatal().Err(err).Msg("cannot build the viewer")
	}
	ctx := context.Background()
	if err := app.Start(ctx); err != nil {
		log.Fatal().Err(err).Msg("cannot start the viewer")
	}
	win.ShowAndRun()
	if err := app.Stop(ctx); err != nil {
		log.Error().Err(err).Msg("did not stop cleanly")
	}
}

func provideConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}
	cfg.SetupLogger()
	return cfg, nil
}

func newFyneApp() fyne.App { return fyneapp.NewWithID("io.ezcards.mrav") }

// The screen and the panel take the Fyne app only so that it exists before their widgets do.
func newScreen(_ fyne.App, svc move.MoveService) *moveui.Screen {
	return moveui.NewScreen(svc, icons.Bytes)
}

func newDemo(_ fyne.App, svc move.MoveService, src *fakemonster.Source) (*demoui.Panel, error) {
	genera, err := svc.ListGenera(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list the genera: %w", err)
	}
	return demoui.New(src, genera), nil
}

func newWindow(a fyne.App, screen *moveui.Screen, demo *demoui.Panel) fyne.Window {
	w := a.NewWindow("mrav - Move Viewer")
	split := container.NewHSplit(screen.Content(), demo.Content())
	split.SetOffset(0.72)
	w.SetContent(split)
	return w
}

// refreshLoop asks the screen to reload every RefreshMS, from a goroutine that lives as long as the app.
func refreshLoop(lc fx.Lifecycle, cfg *config.Config, screen *moveui.Screen) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				ticker := time.NewTicker(time.Duration(cfg.RefreshMS) * time.Millisecond)
				defer ticker.Stop()
				for {
					screen.Load(ctx)
					select {
					case <-ctx.Done():
						return
					case <-ticker.C:
					}
				}
			}()
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}
