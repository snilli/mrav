// Command mrav is the Monster Rancher 2 viewer window.
package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	"app/config"
	"app/internal/core/item"
	"app/internal/core/item/repository/db/itemrepo"
	"app/internal/core/item/service/itemsvc"
	"app/internal/core/move"
	"app/internal/core/move/repository/db/moverepo"
	"app/internal/core/move/service/movesvc"
	"app/internal/core/profile"
	"app/internal/core/profile/repository/db/profilerepo"
	"app/internal/core/profile/service/profilesvc"
	"app/internal/core/training"
	"app/internal/core/training/repository/db/trainingrepo"
	"app/internal/core/training/service/trainingsvc"
	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/itemui"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/routeui"
	"app/internal/delivery/ui/shellui"
	"app/internal/delivery/ui/trainingui"
	"app/internal/infrastructure/database"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
	"app/internal/infrastructure/gamememory"
	"app/internal/infrastructure/icons"
	"app/internal/port"
)

// options is the whole dependency graph. Each repository and service constructor returns its port, so none needs
// fx.Annotate; newSource is the one constructor that gives several types, because the services need only the
// ports of the monster and the player while the demo panel needs the fakes behind them, when there are any.
func options() fx.Option {
	return fx.Options(
		fx.Provide(provideConfig, database.ProvideDatabase),
		fx.Provide(moverepo.New, movesvc.New),
		fx.Provide(profilerepo.New, profilesvc.New),
		fx.Provide(trainingrepo.New, trainingsvc.New),
		fx.Provide(itemrepo.New, itemsvc.New),
		fx.Provide(newSource),
		fx.Provide(newFyneApp, newParts, newDemo, newWindow),
		fx.Invoke(refreshLoops),
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

// canReadGames is whether this system can read a running game's memory, which only Windows can.
var canReadGames = runtime.GOOS == "windows"

// newSource is where the monster and the player come from: a running game's memory, one source for both, or the
// demo panel's hand-set fakes, which are the last two values and nil unless the demo is shown. With auto it is
// whichever game is running, and the demo where no game can be read.
func newSource(cfg *config.Config) (port.MonsterSource, port.PlayerSource, *fakemonster.Source, *fakeplayer.Source) {
	switch {
	case cfg.Game == "duckstation":
		game := gamememory.DuckStation()
		return game, game, nil, nil
	case cfg.Game == "mr2dx":
		game := gamememory.MR2DX()
		return game, game, nil, nil
	case cfg.Game == "auto" && canReadGames:
		game := gamememory.AutoDetect()
		return game, game, nil, nil
	}
	monsters, players := fakemonster.New(), fakeplayer.New()
	return monsters, players, monsters, players
}

func newFyneApp() fyne.App {
	a := fyneapp.NewWithID("io.ezcards.mrav")
	a.Settings().SetTheme(shellui.Theme())
	return a
}

// loader is a page that can be told to ask its service for what it shows.
type loader interface{ Load(ctx context.Context) }

// newParts builds the pages of the window, without the demo panel. They take the Fyne app only so that it exists
// before their widgets do. The Route panel shows the map of the trip that the monster is on in a game, and every map
// in the demo, which is when there are fakes to set a monster by.
func newParts(
	_ fyne.App, profiles profile.ProfileService, moves move.MoveService, trainings training.TrainingService,
	items item.ItemService, monsters *fakemonster.Source,
) shellui.Parts {
	return shellui.Parts{
		Dashboard: profileui.NewDashboard(profiles),
		Training:  trainingui.NewScreen(trainings),
		Moves:     moveui.NewScreen(moves, icons.Bytes),
		Items:     itemui.NewScreen(items),
		Route:     routeui.NewScreen(profiles, monsters != nil),
	}
}

// loaders are the pages, which are all asked at the rate in the config.
func loaders(p shellui.Parts) []loader {
	return []loader{p.Dashboard, p.Moves, p.Training, p.Items, p.Route}
}

// newDemo is nil when a game is read, because nobody sets that monster by hand.
func newDemo(_ fyne.App, svc move.MoveService, monsters *fakemonster.Source, players *fakeplayer.Source) (*demoui.Panel, error) {
	if monsters == nil {
		return nil, nil
	}
	genera, err := svc.ListGenera(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list the genera: %w", err)
	}
	panel := demoui.New(monsters, players, genera)
	panel.Randomize() // so that the pages have something to show from the start
	return panel, nil
}

func newWindow(a fyne.App, p shellui.Parts, demo *demoui.Panel) fyne.Window {
	w := a.NewWindow("mrav - Monster Rancher 2 Viewer")
	w.Resize(fyne.NewSize(1500, 900))
	p.Demo = demo
	w.SetContent(shellui.New(p).Content())
	return w
}

// refreshLoops asks the pages to reload at the rate in the config, from a goroutine that lives as long as the app.
func refreshLoops(lc fx.Lifecycle, cfg *config.Config, p shellui.Parts) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				ticker := time.NewTicker(time.Duration(cfg.RefreshMS) * time.Millisecond)
				defer ticker.Stop()
				for {
					for _, s := range loaders(p) {
						s.Load(ctx)
					}
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
