package gamememory

import (
	"debug/macho"
	"errors"
	"fmt"
	"io"
)

const (
	nStab = 0xE0
	nType = 0x0E
	nSect = 0x0E
)

func machoSymbol(r io.ReaderAt, cpu macho.Cpu, name string) (uint64, error) {
	f, err := machoFor(r, cpu)
	if err != nil {
		return 0, err
	}
	var text *macho.Segment
	for _, l := range f.Loads {
		if s, ok := l.(*macho.Segment); ok && s.Name == "__TEXT" {
			text = s
		}
	}
	if text == nil {
		return 0, errors.New("has no __TEXT segment")
	}
	if f.Symtab != nil {
		for _, s := range f.Symtab.Syms {
			if s.Name == name && s.Type&nStab == 0 && s.Type&nType == nSect {
				return s.Value - text.Addr, nil
			}
		}
	}
	return 0, fmt.Errorf("no symbol named %s", name)
}

func machoFor(r io.ReaderAt, cpu macho.Cpu) (*macho.File, error) {
	fat, err := macho.NewFatFile(r)
	if errors.Is(err, macho.ErrNotFat) {
		return macho.NewFile(r)
	}
	if err != nil {
		return nil, err
	}
	for _, a := range fat.Arches {
		if a.Cpu == cpu {
			return a.File, nil
		}
	}
	return nil, fmt.Errorf("no code for %v in it", cpu)
}
