// Package icons serves the pictures of the moves, embedded in the program: the 637 move icons, named
// "<Breed>_<n>", and the overlays the screen draws on them.
package icons

import (
	"embed"
	"fmt"
)

//go:embed assets/*.png
var assets embed.FS

// Bytes is the PNG of an icon by key: a move icon ("Pixie_1") or an overlay ("No_Move", "hilight",
// "skillchecked", "skillget").
func Bytes(key string) ([]byte, error) {
	b, err := assets.ReadFile("assets/" + key + ".png")
	if err != nil {
		return nil, fmt.Errorf("no icon %q", key)
	}
	return b, nil
}
