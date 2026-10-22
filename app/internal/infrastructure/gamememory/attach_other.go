//go:build !windows && !(darwin && cgo)

package gamememory

import "errors"

var errWindowsOnly = errors.New("reading a game needs Windows or macOS")

func attachDuckStation() (Memory, error) { return nil, errWindowsOnly }
func attachMR2DX() (Memory, error)       { return nil, errWindowsOnly }
