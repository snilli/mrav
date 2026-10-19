package gamememory

import (
	"bytes"
	"debug/pe"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// exportRVA is where a 64-bit PE file's export table puts name: the offset of the exported thing from the
// base address the module is loaded at. DuckStation exports its pointer to the console's RAM this way.
func exportRVA(r io.ReaderAt, name string) (uint32, error) {
	f, err := pe.NewFile(r)
	if err != nil {
		return 0, err
	}
	oh, ok := f.OptionalHeader.(*pe.OptionalHeader64)
	if !ok {
		return 0, errors.New("not a 64-bit program")
	}
	dir := oh.DataDirectory[pe.IMAGE_DIRECTORY_ENTRY_EXPORT].VirtualAddress
	if dir == 0 {
		return 0, errors.New("exports nothing")
	}

	// read is n bytes at an address that is relative to the module's base; the file lays its sections out differently.
	read := func(rva uint32, n int) ([]byte, error) {
		for _, s := range f.Sections {
			if rva >= s.VirtualAddress && rva-s.VirtualAddress < max(s.VirtualSize, s.Size) {
				p := make([]byte, n)
				_, err := s.ReadAt(p, int64(rva-s.VirtualAddress))
				return p, err
			}
		}
		return nil, fmt.Errorf("address %#x is in no section", rva)
	}
	u32 := func(rva uint32) (uint32, error) {
		p, err := read(rva, 4)
		return binary.LittleEndian.Uint32(p), err
	}

	// The directory is 40 bytes: the number of names at 24, then where the function, name and ordinal tables are.
	head, err := read(dir, 40)
	if err != nil {
		return 0, err
	}
	count, funcs := binary.LittleEndian.Uint32(head[24:]), binary.LittleEndian.Uint32(head[28:])
	nameTable, ordTable := binary.LittleEndian.Uint32(head[32:]), binary.LittleEndian.Uint32(head[36:])

	want := append([]byte(name), 0)
	for i := range count {
		at, err := u32(nameTable + 4*i)
		if err != nil {
			return 0, err
		}
		if got, err := read(at, len(want)); err != nil || !bytes.Equal(got, want) {
			continue
		}
		o, err := read(ordTable+2*i, 2)
		if err != nil {
			return 0, err
		}
		return u32(funcs + 4*uint32(binary.LittleEndian.Uint16(o)))
	}
	return 0, fmt.Errorf("no export named %s", name)
}
