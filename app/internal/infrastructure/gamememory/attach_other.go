//go:build !windows

package gamememory

import "errors"

var errWindowsOnly = errors.New("reading a game needs Windows")

func attachDuckStation() (Memory, error) { return nil, errWindowsOnly }
func attachMR2DX() (Memory, error)       { return nil, errWindowsOnly }
