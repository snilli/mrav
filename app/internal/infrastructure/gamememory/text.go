package gamememory

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

//go:embed steam_glyphs.tsv
var steamGlyphsTSV string

var steamGlyphs = parseGlyphs(steamGlyphsTSV)

func parseGlyphs(tsv string) map[int]string {
	glyphs := map[int]string{}
	for line := range strings.Lines(tsv) {
		code, text, ok := strings.Cut(strings.TrimRight(line, "\r\n"), "\t")
		n, err := strconv.ParseUint(code, 16, 16)
		if !ok || err != nil {
			panic(fmt.Sprintf("gamememory: bad glyph line %q", line))
		}
		glyphs[int(n)] = text
	}
	return glyphs
}

func readText(r *reader, t Text) string {
	if t.Addr == 0 {
		return ""
	}
	b := r.read(t.Addr, t.Len+1)
	var name strings.Builder
	for i := 0; i < t.Len; i++ {
		code := int(b[i])
		if !t.Steam {
			if code == 255 {
				break
			}
			name.WriteString(ps1Glyphs[min(code, len(ps1Glyphs)-1)])
			continue
		}
		switch {
		case code >= 0xB0 && code <= 0xB6:
			code = code<<8 | int(b[i+1])
			i++
		case code == 0xFF:
			return name.String()
		}
		if glyph, ok := steamGlyphs[code]; ok {
			name.WriteString(glyph)
		} else {
			name.WriteString("?")
		}
	}
	return name.String()
}
