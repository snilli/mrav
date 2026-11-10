package domain

type Current struct {
	Settings  Settings
	Reactions Reactions
	Effects   Effects
	Date      int
	Dated     bool
	Who       Who
	Acting    bool
}
