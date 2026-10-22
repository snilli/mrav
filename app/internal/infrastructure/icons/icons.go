package icons

import (
	"embed"
	"fmt"
)

//go:embed assets/*.png
var assets embed.FS

func Bytes(key string) ([]byte, error) {
	b, err := assets.ReadFile("assets/" + key + ".png")
	if err != nil {
		return nil, fmt.Errorf("no icon %q", key)
	}
	return b, nil
}
