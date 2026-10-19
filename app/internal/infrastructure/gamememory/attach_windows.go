//go:build windows

package gamememory

// Written from MR2AV's attach code and Microsoft's documentation, and only compiled so far (GOOS=windows go vet):
// it has not been run against a live DuckStation or MR2DX.

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

// steamRAMOffset is where the English release of MR2DX keeps the console's RAM inside MF2.exe. The Japanese
// release keeps it at 0x2CA504 and is not supported: MR2AV tells the two apart by the exe's file description.
const steamRAMOffset = 0x2DEC6C

// process is a running program whose memory can be read.
type process struct {
	handle windows.Handle
	base   uintptr // where its exe is loaded
	path   string  // the exe's file
}

// openProcess finds a running program whose exe name starts with prefix, in lower case, as MR2AV does.
// name is what to call it when it is not running.
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
	if errors.Is(err, windows.ERROR_NO_MORE_FILES) { // the end of the list, not a failure
		err = nil
	}
	return 0, false, err
}

// mainModule is where a process's exe is loaded, and the file it was loaded from.
func mainModule(pid uint32) (base uintptr, path string, err error) {
	snap, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE|windows.TH32CS_SNAPMODULE32, pid)
	if err != nil {
		return 0, "", err
	}
	defer windows.CloseHandle(snap)

	m := windows.ModuleEntry32{Size: uint32(unsafe.Sizeof(windows.ModuleEntry32{}))}
	if err := windows.Module32First(snap, &m); err != nil { // the first module is the exe
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

// processMemory is the console's RAM, which starts at ram inside the process.
type processMemory struct {
	*process
	ram uintptr
}

func (m processMemory) Read(addr uint32, p []byte) error {
	return m.process.read(m.ram+uintptr(addr), p)
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

	// DuckStation has no fixed place for the RAM: its exe exports a pointer to it, named RAM. The export is
	// in the exe's header on disk; the pointer itself is read from the running program.
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
