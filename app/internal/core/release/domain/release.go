package domain

import (
	"regexp"
	"strconv"
)

type Release struct {
	Tag, URL string
}

var tag = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)$`)

func parse(v string) ([3]int, bool) {
	var out [3]int
	m := tag.FindStringSubmatch(v)
	if m == nil {
		return out, false
	}
	for i := range out {
		out[i], _ = strconv.Atoi(m[i+1])
	}
	return out, true
}

func Newer(running, latest string) bool {
	have, ok := parse(running)
	if !ok {
		return false
	}
	got, ok := parse(latest)
	if !ok {
		return false
	}
	for i := range have {
		if got[i] != have[i] {
			return got[i] > have[i]
		}
	}
	return false
}

func IsRelease(version string) bool {
	_, ok := parse(version)
	return ok
}
