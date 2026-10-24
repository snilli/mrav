# mrav

Go + Fyne viewer for Monster Rancher 2, ported from MR2AV. It reads the running game's RAM: DuckStation (PlayStation) on macOS and Windows, MR2DX (Steam) on Windows. `GAME=demo` needs no game.

## Layout

- `go.work` joins `app` (the program) and `orm` (Ent schemas).
- `app/cmd/mrav` wires everything with fx. `app/internal`: `core/<name>/{domain,service,repository}` (rules live in `domain`), `infrastructure/{gamememory,database,memo}`, `delivery/ui/<name>ui` (Fyne). Mocks are in `app/mock` (`make mock`).
- `data/mrav.sqlite` is the shipped database. The app never migrates it: a schema change means Ent plus a hand edit of this file, and a test fails if its indexes differ from Ent's. Do not run ANALYZE on it. `moves` holds a copy of every move per game version (`mr2` PlayStation, `mr2dx` Steam), picked by `gamememory.Layout.Version` through `Monster.Version`.
- `docs/` and `dist/` are local only (`.git/info/exclude`). Findings and decisions are in `docs/pages-notes.md`.

## Commands

`make dev`, `make mac-dev` (signs with `app/debugger.entitlements`; macOS refuses to read DuckStation otherwise), `make watch` (air, signed like `mac-dev`: rebuilds and reopens the viewer on a `.go` change or a write to `data/mrav.sqlite`; the way to show the user the app), `make build`, `make build-mac`, `make test` (ginkgo), `make test-integration`, `make generate` (Ent), `make mock`. Windows: `app/build-windows.sh` (mingw-w64). Build tags: `embeddb`, `no_emoji`.

## Rules for a change

- Tests first, and watch them fail. Specs are Ginkgo and Gomega.
- After any Go change, from `app/`: `gofmt -l -w .`, `go fix ./...`, `go vet ./...`, `GOOS=windows CGO_ENABLED=0 go vet ./internal/infrastructure/gamememory/`, `go test ./...` (and gofmt and vet in `orm/` when it changed).
- `go test` caches a result that depends on `data/mrav.sqlite`: after changing that file run `go test ./... -count=1`.
- No comments in code. At most one short line where the code hides a constraint that would bite the next editor. Credits and sources go in the README, not in comments.
- Rules and numbers belong in the domain; presenters stay thin. Every query goes through an index and is read once (`infrastructure/memo`).
- Fyne: `Show()` only clears Hidden, so call `Refresh()` on the parent after it.
- Commit messages are short, and never carry a `Co-Authored-By` line or any other mention of Claude. Do not commit, push or release unless asked.

## Reading the game

- Addresses live in `app/internal/infrastructure/gamememory/layout.go`. The tests type the same addresses again from their sources (MR2AV, LegendCup's GameShark page), so a wrong address cannot pass quietly.
- Say what was seen in the game and what was only inferred from another layout.
- Testing by writing the emulator's RAM is allowed on a test game, never saved to the memory card. Log the old value before each write.
