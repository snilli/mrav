# Changelog

## v0.5.2

- **Nature**: the bar shows how far nature can go from its base (base minus 100 to base plus 100, never past -100 or
  100) as a lighter band, the base as a thin white line, and fills from the base to the current nature, red below 0
  and green above. Hover it for the min, base and max. The Base Nature line is gone.
- **Easier to read**: grey boxes ("Bad" nature, Silver Peach, the Ranch altar) have black text now, not white.

## v0.5.1

- **New version notice**: when a newer release is out, a bar at the top of the window says so. Download opens the
  release page; the app does not download or replace itself. It asks GitHub once at start and gives up after 3
  seconds; with no internet it just carries on.
- **Route**: the item finder no longer says "needs SKI 0". It says which SKI range finds the item, or that the
  monster's nature never finds it there, and when a monster of the other nature would find it at this SKI.
- **Calendar**: tap a colour to show or hide that kind of entry; the choice is remembered between runs.
- **Planner**: follows the loads, food, items and actions you actually do in the game, and reads the drill stars from
  the game. It only acts on what it saw this week, and counts weeks done by the game date.
- **MR2DX**: knows the 256 breeds added by the More Monsters mod. Not yet tried with the mod running.
- **Windows**: only `mrav.exe` now; the UPX build (`mrav-upx.exe`) is gone.
- **Credits**: LegendCup is MonsterFenrick's; thanks to Teawch.

## v0.5.0

First release with the README. Earlier changes are in the git history.
