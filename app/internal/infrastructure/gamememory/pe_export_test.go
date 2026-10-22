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

func buildPE(exports []export) []byte {
	const (
		peAt      = 0x40
		optSize   = 240
		sectionAt = peAt + 4 + 20 + optSize
		dataAt    = 0x200
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
	put([]uint32{0, 0})
	put([]uint16{0, 0})
	put([]uint32{stringsRVA, 1, n, n, funcsRVA, namesRVA, ordsRVA})
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
	le.PutUint16(coff[0:], 0x8664)
	le.PutUint16(coff[2:], 1)
	le.PutUint16(coff[16:], optSize)
	opt := file[peAt+4+20:]
	le.PutUint16(opt[0:], 0x20B)
	le.PutUint32(opt[108:], 16)
	le.PutUint32(opt[112:], dataRVA)
	le.PutUint32(opt[116:], uint32(data.Len()))
	sec := file[sectionAt:]
	copy(sec, ".edata")
	le.PutUint32(sec[8:], 0x200)
	le.PutUint32(sec[12:], dataRVA)
	le.PutUint32(sec[16:], 0x200)
	le.PutUint32(sec[20:], dataAt)
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
