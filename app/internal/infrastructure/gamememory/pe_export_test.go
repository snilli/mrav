package gamememory

import (
	"bytes"
	"encoding/binary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type export struct {
	name string
	rva  uint32
}

// buildPE is the smallest 64-bit PE file that has an export table: the headers, and one section, .edata,
// holding the table. The function addresses are in the reverse order of the names, so that finding an export
// has to go through the ordinals.
func buildPE(exports []export) []byte {
	const (
		peAt      = 0x40 // where the PE signature is: the DOS header's e_lfanew
		optSize   = 240  // a PE32+ optional header with its 16 data directories
		sectionAt = peAt + 4 + 20 + optSize
		dataAt    = 0x200 // the section's place in the file
		dataRVA   = 0x1000
		dirSize   = 40
	)
	n := uint32(len(exports))
	funcsRVA := uint32(dataRVA + dirSize)
	namesRVA := funcsRVA + 4*n
	ordsRVA := namesRVA + 4*n
	stringsRVA := ordsRVA + 2*n

	var data bytes.Buffer
	put := func(v any) { Expect(binary.Write(&data, binary.LittleEndian, v)).To(Succeed()) }
	put([]uint32{0, 0})                                             // Characteristics, TimeDateStamp
	put([]uint16{0, 0})                                             // MajorVersion, MinorVersion
	put([]uint32{stringsRVA, 1, n, n, funcsRVA, namesRVA, ordsRVA}) // Name, Base, NumberOfFunctions, NumberOfNames, the three tables
	for i := range exports {
		put(exports[n-1-uint32(i)].rva)
	}
	next := stringsRVA
	for _, e := range exports {
		put(next)
		next += uint32(len(e.name)) + 1
	}
	for i := range exports {
		put(uint16(n - 1 - uint32(i)))
	}
	for _, e := range exports {
		data.WriteString(e.name)
		data.WriteByte(0)
	}

	le := binary.LittleEndian
	file := make([]byte, dataAt+0x200)
	copy(file, "MZ")
	le.PutUint32(file[0x3C:], peAt)
	copy(file[peAt:], "PE\x00\x00")
	coff := file[peAt+4:]
	le.PutUint16(coff[0:], 0x8664)   // Machine: x64
	le.PutUint16(coff[2:], 1)        // NumberOfSections
	le.PutUint16(coff[16:], optSize) // SizeOfOptionalHeader
	opt := file[peAt+4+20:]
	le.PutUint16(opt[0:], 0x20B)                // Magic: PE32+
	le.PutUint32(opt[108:], 16)                 // NumberOfRvaAndSizes
	le.PutUint32(opt[112:], dataRVA)            // the export directory: where,
	le.PutUint32(opt[116:], uint32(data.Len())) // and how long
	sec := file[sectionAt:]
	copy(sec, ".edata")
	le.PutUint32(sec[8:], 0x200)    // VirtualSize
	le.PutUint32(sec[12:], dataRVA) // VirtualAddress
	le.PutUint32(sec[16:], 0x200)   // SizeOfRawData
	le.PutUint32(sec[20:], dataAt)  // PointerToRawData
	copy(file[dataAt:], data.Bytes())
	return file
}

var _ = Describe("exportRVA", func() {
	It("finds an export by its name, whatever its ordinal", func() {
		pe := buildPE([]export{{"Zeta", 0x2000}, {"RAM", 0x3A40}})

		rva, err := exportRVA(bytes.NewReader(pe), "RAM")
		Expect(err).NotTo(HaveOccurred())
		Expect(rva).To(Equal(uint32(0x3A40)))

		rva, err = exportRVA(bytes.NewReader(pe), "Zeta")
		Expect(err).NotTo(HaveOccurred())
		Expect(rva).To(Equal(uint32(0x2000)))
	})

	It("does not take a longer name that starts the same for the one it wants", func() {
		_, err := exportRVA(bytes.NewReader(buildPE([]export{{"RAMS", 0x2000}})), "RAM")
		Expect(err).To(MatchError(ContainSubstring("RAM")))
	})

	It("says when the file has no such export", func() {
		_, err := exportRVA(bytes.NewReader(buildPE([]export{{"Zeta", 0x2000}})), "RAM")
		Expect(err).To(MatchError(ContainSubstring("RAM")))
	})

	It("refuses what is not a PE file", func() {
		_, err := exportRVA(bytes.NewReader([]byte("not a program")), "RAM")
		Expect(err).To(HaveOccurred())
	})
})
