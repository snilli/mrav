# mrav

An all-in-one helper for Monster Rancher. Leave it open next to the game and it shows your monster's stats, what each
drill will give and how likely it is to go great or get cheated, which moves an errantry can teach, the best
expedition route, what's on the calendar, and a planner that follows along as you play.

It's a Go port of Lexichu's [MR2AV](https://github.com/Lexichu/mr2av_repo).

I've played Monster Rancher 1 and 2 since I was a kid on the PlayStation, and I still do. MR2AV was the viewer I wanted,
but I wanted one app that does it all: read the game, work out the drills, plan the weeks, find the route, no web
calculators open on the side. So I rewrote it in Go as a single file you download and run.

Works with:

- MR2 (English, PlayStation) in DuckStation, on Windows and macOS
- Monster Rancher 1 & 2 DX on Steam, Windows only

## Download

Grab `mrav.exe` (Windows) or `mrav-mac-arm64.dmg` (Mac) from [Releases](https://github.com/snilli/mrav/releases),
start the game, then open it. With no game running you get a demo monster to play with.

On Windows there's also the smaller `mrav-upx.exe`; if your antivirus complains about it, use `mrav.exe`.

On a Mac, open the dmg and drag mrav into Applications. The app isn't from a registered Apple developer, so the
first time macOS will refuse to open it: right-click it and choose Open, or go to System Settings > Privacy &
Security and click Open Anyway.

## Build

Go, Fyne, DDD + clean architecture, one binary with the database built in. Needs Go and a C compiler.

```sh
make dev               # run
make mac-dev           # run on macOS (signed, so it's allowed to read DuckStation)
make build-mac         # dist/mrav.app and dist/mrav-mac-arm64.dmg
make build-windows     # dist/mrav.exe
make test
```

## Credits

- [MR2AV](https://github.com/Lexichu/mr2av_repo) by Lexichu: memory addresses, names, item list, and the original app
- [LegendCup](https://legendcup.com) by Teawch: moves, breeds, drill data, the training planner (with MonsterFenrick),
  errantry chances, expedition maps, and the GameShark page the PlayStation ranch addresses come from. Much of it
  comes from SmilingFaces96.
- The MR2DX Cheat Engine table on [fearlessrevolution](https://fearlessrevolution.com/viewtopic.php?t=18453): the
  extra Steam addresses
