package gamememory

import (
	"bytes"
	"debug/macho"
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type symbol struct {
	name  string
	value uint64
}

func buildMachO(cpu macho.Cpu, textAddr uint64, symbols []symbol, stabs ...symbol) []byte {
	const (
		headerSize  = 32
		segmentSize = 72
		symtabSize  = 24
		nlistSize   = 16
	)
	type entry struct {
		symbol
		typ byte
	}
	var entries []entry
	for _, s := range stabs {
		entries = append(entries, entry{s, 0x20})
	}
	for i := range symbols {
		entries = append(entries, entry{symbols[len(symbols)-1-i], 0x0F})
	}
	symoff := uint32(headerSize + segmentSize + symtabSize)
	stroff := symoff + nlistSize*uint32(len(entries))

	var strtab bytes.Buffer
	strtab.WriteByte(0)
	var nlists bytes.Buffer
	for _, e := range entries {
		Expect(binary.Write(&nlists, binary.LittleEndian, uint32(strtab.Len()))).To(Succeed())
		nlists.Write([]byte{e.typ, 1, 0, 0})
		Expect(binary.Write(&nlists, binary.LittleEndian, e.value)).To(Succeed())
		strtab.WriteString(e.name)
		strtab.WriteByte(0)
	}

	var b bytes.Buffer
	put := func(v any) { Expect(binary.Write(&b, binary.LittleEndian, v)).To(Succeed()) }
	put([]uint32{0xFEEDFACF, uint32(cpu), 0, 2, 2, segmentSize + symtabSize, 0, 0})
	put([]uint32{0x19, segmentSize})
	b.WriteString("__TEXT\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00")
	put([]uint64{textAddr, 0x4000, 0, 0})
	put([]uint32{5, 5, 0, 0})
	put([]uint32{0x2, symtabSize})
	put([]uint32{symoff, uint32(len(entries)), stroff, uint32(strtab.Len())})
	b.Write(nlists.Bytes())
	b.Write(strtab.Bytes())
	return b.Bytes()
}

func buildFat(cpus []macho.Cpu, slices [][]byte) []byte {
	var b bytes.Buffer
	put := func(v any) { Expect(binary.Write(&b, binary.BigEndian, v)).To(Succeed()) }
	headerLen := uint32(8 + 20*len(slices))
	put([]uint32{0xCAFEBABE, uint32(len(slices))})
	var body bytes.Buffer
	for i, s := range slices {
		at := (headerLen + uint32(body.Len()) + 15) &^ 15
		put([]uint32{uint32(cpus[i]), 0, at, uint32(len(s)), 4})
		body.Write(make([]byte, at-headerLen-uint32(body.Len())))
		body.Write(s)
	}
	b.Write(body.Bytes())
	return b.Bytes()
}

var _ = Describe("machoSymbol", func() {
	const textAt = 0x100000000

	It("is the symbol's distance from the start of __TEXT", func() {
		file := buildMachO(macho.CpuArm64, textAt, []symbol{{"_main", textAt + 0x1000}, {"_RAM", textAt + 0x810480}})

		off, err := machoSymbol(bytes.NewReader(file), macho.CpuArm64, "_RAM")

		Expect(err).NotTo(HaveOccurred())
		Expect(off).To(BeNumerically("==", 0x810480))
	})

	It("is not the symbol of another name that starts the same", func() {
		file := buildMachO(macho.CpuArm64, textAt, []symbol{{"_RAMP", textAt + 0x20}, {"_RAM", textAt + 0x30}})

		off, err := machoSymbol(bytes.NewReader(file), macho.CpuArm64, "_RAM")

		Expect(err).NotTo(HaveOccurred())
		Expect(off).To(BeNumerically("==", 0x30))
	})

	It("is not a debugger's entry of the same name, which nm does not show", func() {
		file := buildMachO(macho.CpuArm64, textAt, []symbol{{"_RAM", textAt + 0x810480}}, symbol{"_RAM", 0})

		off, err := machoSymbol(bytes.NewReader(file), macho.CpuArm64, "_RAM")

		Expect(err).NotTo(HaveOccurred())
		Expect(off).To(BeNumerically("==", 0x810480))
	})

	It("reads the slice of the cpu that the program runs on, in a universal file", func() {
		amd := buildMachO(macho.CpuAmd64, textAt, []symbol{{"_RAM", textAt + 0x700000}})
		arm := buildMachO(macho.CpuArm64, textAt, []symbol{{"_RAM", textAt + 0x810480}})
		file := buildFat([]macho.Cpu{macho.CpuAmd64, macho.CpuArm64}, [][]byte{amd, arm})

		fromArm, err := machoSymbol(bytes.NewReader(file), macho.CpuArm64, "_RAM")
		Expect(err).NotTo(HaveOccurred())
		fromAmd, err := machoSymbol(bytes.NewReader(file), macho.CpuAmd64, "_RAM")
		Expect(err).NotTo(HaveOccurred())

		Expect(fromArm).To(BeNumerically("==", 0x810480))
		Expect(fromAmd).To(BeNumerically("==", 0x700000))
	})

	It("says so when a universal file has no slice for the cpu", func() {
		file := buildFat([]macho.Cpu{macho.CpuAmd64}, [][]byte{buildMachO(macho.CpuAmd64, textAt, []symbol{{"_RAM", textAt}})})

		_, err := machoSymbol(bytes.NewReader(file), macho.CpuArm64, "_RAM")

		Expect(err).To(MatchError(ContainSubstring("no code for")))
	})

	It("says so when the program has no such symbol", func() {
		file := buildMachO(macho.CpuArm64, textAt, []symbol{{"_main", textAt}})

		_, err := machoSymbol(bytes.NewReader(file), macho.CpuArm64, "_RAM")

		Expect(err).To(MatchError(ContainSubstring("no symbol named _RAM")))
	})

	It("says so when the file is no Mach-O", func() {
		_, err := machoSymbol(bytes.NewReader([]byte("MZ not a mac program at all")), macho.CpuArm64, "_RAM")

		Expect(err).To(HaveOccurred())
	})
})
