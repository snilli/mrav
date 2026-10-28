package main

import (
	"context"
	"fmt"
	"image/color"
	"runtime"
	"time"

	"fyne.io/fyne/v2"
	fyneapp "fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"github.com/rs/zerolog/log"
	"go.uber.org/fx"

	"app/config"
	"app/internal/core/drill"
	"app/internal/core/drill/repository/db/drillrepo"
	"app/internal/core/drill/service/drillsvc"
	"app/internal/core/item"
	"app/internal/core/item/repository/db/itemrepo"
	"app/internal/core/item/service/itemsvc"
	"app/internal/core/move"
	"app/internal/core/move/repository/db/moverepo"
	"app/internal/core/move/service/movesvc"
	"app/internal/core/planner"
	"app/internal/core/planner/repository/db/plannerrepo"
	"app/internal/core/planner/service/plannersvc"
	"app/internal/core/profile"
	"app/internal/core/profile/repository/db/profilerepo"
	"app/internal/core/profile/service/profilesvc"
	"app/internal/delivery/ui/demoui"
	"app/internal/delivery/ui/drillui"
	"app/internal/delivery/ui/errantryui"
	"app/internal/delivery/ui/itemui"
	"app/internal/delivery/ui/moveui"
	"app/internal/delivery/ui/plannerui"
	"app/internal/delivery/ui/profileui"
	"app/internal/delivery/ui/routeui"
	"app/internal/delivery/ui/shellui"
	"app/internal/infrastructure/database"
	"app/internal/infrastructure/fakemonster"
	"app/internal/infrastructure/fakeplayer"
	"app/internal/infrastructure/gamememory"
	"app/internal/infrastructure/icons"
	"app/internal/port"
)

func options() fx.Option {
	return fx.Options(
		fx.Provide(provideConfig, database.ProvideDatabase),
		fx.Provide(moverepo.New, movesvc.New),
		fx.Provide(profilerepo.New, profilesvc.New),
		fx.Provide(drillrepo.New, drillsvc.New),
		fx.Provide(plannerrepo.New, plannersvc.New),
		fx.Provide(itemrepo.New, itemsvc.New),
		fx.Provide(newSource),
		fx.Provide(newFyneApp, newParts, newDemo, newWindow),
		fx.Invoke(refreshLoops),
	)
}

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

var version = "dev"

var canReadGames = runtime.GOOS == "windows" || runtime.GOOS == "darwin"

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
	a := fyneapp.New()
	a.Settings().SetTheme(shellui.Theme())
	return a
}

type loader interface{ Load(ctx context.Context) }

func newParts(
	_ fyne.App, profiles profile.ProfileService, moves move.MoveService,
	drills drill.DrillService, items item.ItemService, planners planner.PlannerService, monsters *fakemonster.Source,
) shellui.Parts {
	return shellui.Parts{
		Dashboard: profileui.NewDashboard(profiles),
		Drills:    drillui.NewScreen(drills),
		Planner:   plannerui.NewScreen(planners),
		Moves:     moveui.NewScreen(moves, icons.Bytes),
		Items:     itemui.NewScreen(items),
		Route:     routeui.NewScreen(profiles),
		Errantry:  errantryui.NewScreen(profiles, moves, moveui.NewIcons(icons.Bytes)),
	}
}

func loaders(p shellui.Parts) []loader {
	return []loader{p.Dashboard, p.Moves, p.Drills, p.Items, p.Route, p.Errantry, p.Planner}
}

func newDemo(_ fyne.App, svc move.MoveService, monsters *fakemonster.Source, players *fakeplayer.Source) (*demoui.Panel, error) {
	if monsters == nil {
		return nil, nil
	}
	genera, err := svc.ListGenera(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list the genera: %w", err)
	}
	panel := demoui.New(monsters, players, genera)
	panel.Randomize()
	return panel, nil
}

var (
	startSize = fyne.NewSize(480, 800)
	minHeight = float32(640)
)

func newWindow(a fyne.App, p shellui.Parts, demo *demoui.Panel) fyne.Window {
	w := a.NewWindow("mrav " + version + " - Monster Rancher 2 Viewer")
	p.Demo = demo
	floor := canvas.NewRectangle(color.Transparent)
	floor.SetMinSize(fyne.NewSize(0, minHeight))
	w.SetContent(container.NewStack(floor, shellui.New(w, p).Content()))
	w.Resize(startSize)
	return w
}

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
