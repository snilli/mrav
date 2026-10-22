//go:build darwin && cgo

package gamememory

import (
	"debug/macho"
	"encoding/binary"
	"os"
	"runtime"
	"strings"
	"unsafe"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("openProcess", func() {
	var p *process

	BeforeEach(func() {
		var err error
		p, err = openProcess(os.Getpid())
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(p.close)
	})

	It("finds where the exe is loaded, which starts with the header of an exe", func() {
		head := make([]byte, 16)
		Expect(p.read(p.base, head)).To(Succeed())

		Expect(binary.LittleEndian.Uint32(head)).To(BeEquivalentTo(macho.Magic64))
		Expect(binary.LittleEndian.Uint32(head[12:])).To(BeEquivalentTo(macho.TypeExec))
	})

	It("reads what the program holds at an address", func() {
		held := []byte("the console's RAM")
		got := make([]byte, len(held))

		Expect(p.read(uintptr(unsafe.Pointer(&held[0])), got)).To(Succeed())

		runtime.KeepAlive(held)
		Expect(got).To(Equal(held))
	})

	It("says so when the address is not the program's", func() {
		Expect(p.read(8, make([]byte, 8))).To(HaveOccurred())
	})

	It("is where the exe's file says its symbols are", func() {
		exe, err := os.Executable()
		Expect(err).NotTo(HaveOccurred())
		f, err := os.Open(exe)
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(f.Close)
		cpu := macho.CpuArm64
		if runtime.GOARCH != "arm64" {
			cpu = macho.CpuAmd64
		}
		thin, err := macho.NewFile(f)
		Expect(err).NotTo(HaveOccurred())
		var name string
		for _, s := range thin.Symtab.Syms {
			if strings.HasSuffix(s.Name, "_Cfunc_close_task") {
				name = s.Name
			}
		}
		Expect(name).NotTo(BeEmpty())
		offset, err := machoSymbol(f, cpu, name)
		Expect(err).NotTo(HaveOccurred())

		inFile, inMemory := make([]byte, 32), make([]byte, 32)
		_, err = f.ReadAt(inFile, int64(offset))
		Expect(err).NotTo(HaveOccurred())
		Expect(p.read(p.base+uintptr(offset), inMemory)).To(Succeed())

		Expect(inMemory).To(Equal(inFile))
	})
})
