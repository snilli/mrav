//go:build windows

package gamememory

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

const steamRAMOffset = 0x2DEC6C

type process struct {
	handle windows.Handle
	base   uintptr
	path   string
}

func openProcess(prefix, name string) (*process, error) {
	pid, found, err := findPID(prefix)
	if err != nil {
		return nil, fmt.Errorf("cannot list the running programs: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("%s is not running", name)
	}
	base, path, err := mainModule(pid)
	if err != nil {
		return nil, fmt.Errorf("cannot look into %s: %w", name, err)
	}
	handle, err := windows.OpenProcess(windows.PROCESS_VM_READ, false, pid)
	if err != nil {
		return nil, fmt.Errorf("cannot open %s to read it: %w", name, err)
	}
	return &process{handle: handle, base: base, path: path}, nil
}

func findPID(prefix string) (pid uint32, found bool, err error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, false, err
	}
	defer windows.CloseHandle(snap)

	e := windows.ProcessEntry32{Size: uint32(unsafe.Sizeof(windows.ProcessEntry32{}))}
	for err = windows.Process32First(snap, &e); err == nil; err = windows.Process32Next(snap, &e) {
		if strings.HasPrefix(strings.ToLower(windows.UTF16ToString(e.ExeFile[:])), prefix) {
			return e.ProcessID, true, nil
		}
	}
	if errors.Is(err, windows.ERROR_NO_MORE_FILES) {
		err = nil
	}
	return 0, false, err
}

func mainModule(pid uint32) (base uintptr, path string, err error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32, pid)
	if err != nil {
		return 0, "", err
	}
	defer windows.CloseHandle(snap)

	m := windows.ModuleEntry32{Size: uint32(unsafe.Sizeof(windows.ModuleEntry32{}))}
	if err := windows.Module32First(snap, &m); err != nil {
		return 0, "", err
	}
	return m.ModBaseAddr, windows.UTF16ToString(m.ExePath[:]), nil
}

func (p *process) read(addr uintptr, buf []byte) error {
	var n uintptr
	if err := windows.ReadProcessMemory(p.handle, addr, &buf[0], uintptr(len(buf)), &n); err != nil {
		return err
	}
	if int(n) != len(buf) {
		return io.ErrUnexpectedEOF
	}
	return nil
}

type processMemory struct {
	*process
	ram uintptr
}

func (m processMemory) Read(addr uint32, p []byte) error {
	return m.process.read(m.ram+uintptr(addr), p)
}

func (m processMemory) ReadModule(offset uint32, p []byte) error {
	return m.process.read(m.process.base+uintptr(offset), p)
}

func (m processMemory) Close() error { return windows.CloseHandle(m.handle) }

func attachMR2DX() (Memory, error) {
	p, err := openProcess("mf2", "Monster Rancher 1&2 DX")
	if err != nil {
		return nil, err
	}
	return processMemory{p, p.base + steamRAMOffset}, nil
}

func attachDuckStation() (_ Memory, err error) {
	p, err := openProcess("duckstation", "DuckStation")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			_ = windows.CloseHandle(p.handle)
		}
	}()

	rva, err := exportOf(p.path, "RAM")
	if err != nil {
		return nil, fmt.Errorf("cannot find DuckStation's RAM: %w", err)
	}
	var ptr [8]byte
	if err := p.read(p.base+uintptr(rva), ptr[:]); err != nil {
		return nil, fmt.Errorf("cannot read DuckStation's RAM pointer: %w", err)
	}
	ram := uintptr(binary.LittleEndian.Uint64(ptr[:]))
	if ram == 0 {
		return nil, errors.New("DuckStation is not running a game")
	}
	return processMemory{p, ram}, nil
}

func exportOf(path, name string) (uint32, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()
	return exportRVA(f, name)
}
