//go:build darwin && cgo

package gamememory

/*
#include <libproc.h>
#include <mach/mach.h>
#include <mach/mach_vm.h>
#include <mach-o/loader.h>
#include <stdlib.h>
#include <string.h>
#include <strings.h>

static pid_t find_pid(const char *prefix) {
	int bytes = proc_listpids(PROC_ALL_PIDS, 0, NULL, 0);
	if (bytes <= 0) return 0;
	bytes += 64 * sizeof(pid_t);
	pid_t *pids = malloc(bytes);
	if (!pids) return 0;
	bytes = proc_listpids(PROC_ALL_PIDS, 0, pids, bytes);
	pid_t found = 0;
	for (int i = 0; i < bytes / (int)sizeof(pid_t) && !found; i++) {
		char name[64];
		if (pids[i] > 0 && proc_name(pids[i], name, sizeof name) > 0 && strncasecmp(name, prefix, strlen(prefix)) == 0)
			found = pids[i];
	}
	free(pids);
	return found;
}

static kern_return_t open_task(pid_t pid, task_t *task) { return task_for_pid(mach_task_self(), pid, task); }

static void close_task(task_t task) { mach_port_deallocate(mach_task_self(), task); }

static kern_return_t read_task(task_t task, uint64_t addr, void *buf, uint64_t n, uint64_t *got) {
	mach_vm_size_t out = 0;
	kern_return_t kr = mach_vm_read_overwrite(task, addr, n, (mach_vm_address_t)buf, &out);
	*got = out;
	return kr;
}

static int exe_base(task_t task, uint64_t *base) {
	mach_vm_address_t addr = 0;
	for (;;) {
		mach_vm_size_t size = 0;
		vm_region_basic_info_data_64_t info;
		mach_msg_type_number_t count = VM_REGION_BASIC_INFO_COUNT_64;
		mach_port_t object;
		if (mach_vm_region(task, &addr, &size, VM_REGION_BASIC_INFO_64, (vm_region_info_t)&info, &count, &object) != KERN_SUCCESS || size == 0)
			return -1;
		if (info.protection & VM_PROT_EXECUTE) {
			struct mach_header_64 head;
			mach_vm_size_t got;
			if (mach_vm_read_overwrite(task, addr, sizeof head, (mach_vm_address_t)&head, &got) == KERN_SUCCESS &&
				head.magic == MH_MAGIC_64 && head.filetype == MH_EXECUTE) {
				*base = addr;
				return 0;
			}
		}
		addr += size;
	}
}
*/
import "C"

import (
	"debug/macho"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"unsafe"
)

type process struct {
	task C.task_t
	base uintptr
}

func machError(kr C.kern_return_t) error { return errors.New(C.GoString(C.mach_error_string(kr))) }

func openProcess(pid int) (*process, error) {
	var task C.task_t
	if kr := C.open_task(C.pid_t(pid), &task); kr != C.KERN_SUCCESS {
		return nil, fmt.Errorf("macOS would not open it (%w); only the mrav app, which is signed for that, and the superuser may", machError(kr))
	}
	var base C.uint64_t
	if C.exe_base(task, &base) != 0 {
		C.close_task(task)
		return nil, errors.New("cannot find where its exe is loaded")
	}
	return &process{task: task, base: uintptr(base)}, nil
}

func (p *process) close() { C.close_task(p.task) }

func (p *process) read(addr uintptr, buf []byte) error {
	var got C.uint64_t
	if kr := C.read_task(p.task, C.uint64_t(addr), unsafe.Pointer(&buf[0]), C.uint64_t(len(buf)), &got); kr != C.KERN_SUCCESS {
		return machError(kr)
	}
	if int(got) != len(buf) {
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

func (m processMemory) Close() error {
	m.process.close()
	return nil
}

func exeOf(pid C.pid_t) (string, macho.Cpu, error) {
	path := make([]byte, C.PROC_PIDPATHINFO_MAXSIZE)
	n := C.proc_pidpath(C.int(pid), unsafe.Pointer(&path[0]), C.uint32_t(len(path)))
	if n <= 0 {
		return "", 0, errors.New("cannot tell where it was started from")
	}
	var arch C.struct_proc_archinfo
	if C.proc_pidinfo(C.int(pid), C.PROC_PIDARCHINFO, 0, unsafe.Pointer(&arch), C.int(unsafe.Sizeof(arch))) <= 0 {
		return "", 0, errors.New("cannot tell which cpu it runs on")
	}
	return string(path[:n]), macho.Cpu(uint32(arch.p_cputype)), nil
}

func attachMR2DX() (Memory, error) {
	return nil, errors.New("Monster Rancher 1&2 DX is read only on Windows")
}

func attachDuckStation() (_ Memory, err error) {
	prefix := C.CString("duckstation")
	defer C.free(unsafe.Pointer(prefix))
	pid := C.find_pid(prefix)
	if pid == 0 {
		return nil, errors.New("DuckStation is not running")
	}

	path, cpu, err := exeOf(pid)
	if err != nil {
		return nil, fmt.Errorf("cannot look into DuckStation: %w", err)
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cannot find DuckStation's RAM: %w", err)
	}
	defer f.Close()
	offset, err := machoSymbol(f, cpu, "_RAM")
	if err != nil {
		return nil, fmt.Errorf("cannot find DuckStation's RAM: %w", err)
	}

	p, err := openProcess(int(pid))
	if err != nil {
		return nil, fmt.Errorf("cannot read DuckStation: %w", err)
	}
	defer func() {
		if err != nil {
			p.close()
		}
	}()

	var ptr [8]byte
	if err := p.read(p.base+uintptr(offset), ptr[:]); err != nil {
		return nil, fmt.Errorf("cannot read DuckStation's RAM pointer: %w", err)
	}
	ram := uintptr(binary.LittleEndian.Uint64(ptr[:]))
	if ram == 0 {
		return nil, errors.New("DuckStation is not running a game")
	}
	return processMemory{p, ram}, nil
}
