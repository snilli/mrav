package gamememory

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"
)

// steamGlyphsTSV is MR2DX's character table, from MR2AV's CharMapping: a code in hex, a tab, the text.
//
//go:embed steam_glyphs.tsv
var steamGlyphsTSV string

var steamGlyphs = parseGlyphs(steamGlyphsTSV)

// parseGlyphs reads the table that is built into the program, so a line that is not a code and its text is a
// mistake in the program and not in anything it was given.
func parseGlyphs(tsv string) map[int]string {
	glyphs := map[int]string{}
	for line := range strings.Lines(tsv) {
		code, text, ok := strings.Cut(strings.TrimSuffix(line, "\n"), "\t")
		n, err := strconv.ParseUint(code, 16, 16)
		if !ok || err != nil {
			panic(fmt.Sprintf("gamememory: bad glyph line %q", line))
		}
		glyphs[int(n)] = text
	}
	return glyphs
}

// readText reads a name. The PlayStation's names are one byte a glyph and end at 255. MR2DX's are one byte for a
// space or a digit and two for the rest, the bytes B0 to B6 beginning a pair, and end at FF. What the table does
// not know shows as "_" on the PlayStation and "?" on Steam, as in MR2AV.
func readText(r *reader, t Text) string {
	if t.Addr == 0 {
		return ""
	}
	b := r.read(t.Addr, t.Len+1) // one more than the name, because a pair can begin on its last byte
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
