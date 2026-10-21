package profileui

import (
	"context"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	player "app/internal/core/player/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/status"
)

// StatNames are the stats as MR2AV abbreviates them, by stat.
var StatNames = [monster.StatCount]string{"LIF", "POW", "INT", "SKI", "SPD", "DEF"}

// The widths of the columns of a row of bars: the name, then the bar, then what it says, then a growth.
const (
	nameWidth  = 70
	valueWidth = 100
	chipWidth  = 84
)

// cocoonInfo is what MR2AV tells a breeder about cocooning.
const cocoonInfo = `Cocooning requires a Worm that has:
- Zero fatigue.
- Less than 30 Stress.
- *80+ Loyalty.
- C Rank or lower.
- Lived for at least 4 years, and is on the Ranch when the calendar reaches 'June, Week 4.'

If the 'Cup Jellies Eaten' counter is green, you have fed enough Cup Jellies for your Worm to become a Beaclon.
If the box is not green, but 'Cocoon Ready' is lit, your Worm will cocoon into a random /Worm monster.

If this cocooning information disappears, your Worm is past C Rank, therefore it cannot cocoon any more.

*79 Loyalty is possible, with one of Spoil/Fear being 79 and the other 80.
78 Loyalty is possible with both Spoil and Fear being 79.
'Cocoon Ready' will only light on 80+ Loyalty, despite these two cases.`

// Dashboard is the monster the game shows, on one screen: a header of who it is and how much life it has left, and
// the vitals of its body. It is two parts so that the window can put them where it likes, and one value so that
// they load together.
type Dashboard struct {
	seen status.Seen[domain.Profile] // what apply last drew, so that the same is not drawn again
	svc  profile.ProfileService
	opts Options
	last *domain.Profile // what is shown, to show it again when an option changes

	gutsPerSecond   *widget.Check // beside the guts rate, which it changes
	growthFromZero  *widget.Check // above the growth rates, which it changes
	status          *widget.Label
	headerBody      fyne.CanvasObject
	vitalsBody      fyne.CanvasObject
	whoRow, lifeRow Beside // how the rows of the header are laid out, side by side or stacked
	header          fyne.CanvasObject
	vitals          fyne.CanvasObject

	breed                                         *ToneBox
	givenName, genus, age, stage, lifespan, money *widget.Label
	lifespanBar, lifeIndexBar                     *Meter
	lifeIndex                                     *ToneBox

	stats                                               [monster.StatCount]*widget.Label
	statBars                                            [monster.StatCount]*Meter
	statNames                                           [monster.StatCount]*ColorText // each in the colour of its stat
	growth                                              [monster.StatCount]*ToneBox
	statTotal                                           *widget.Label
	fatigue, stress                                     *widget.Label
	fatigueBar, stressBar, natureGauge                  *Meter
	nature                                              *ToneBox
	loyalty, rank, form, natureBase, lifeType, playType *widget.Label
	prize, nextSale, drug, drugTip                      *widget.Label
	tournament                                          *widget.Label
	tournamentRow                                       fyne.CanvasObject // hidden where the game's year of the first one was not read
	fame, guts, arena, specials, like, dislike          *ToneBox
	goldPeach, silverPeach, itemUsed, errantry          *Flag
	cocoon                                              fyne.CanvasObject
	cupJellies                                          *ToneBox
	cocoonDone                                          *Flag

	date                         *widget.Label // the game's date in the header, where the game keeps one
	fatigueWord, record, history *widget.Label
	ranchBody                    fyne.CanvasObject // the upgrades, hidden where the game keeps no ranch
	upgradeDay                   *widget.Label     // when the game offers the house's and the stable's upgrades
	facilities                   [3]facilityRow
	breeds                       *fyne.Container // the main breeds, lit where unlocked, under a title with the count
	breedsTitle                  *widget.Label
	breedFlags                   [player.BreedCount]*Flag

	ranchBox        fyne.CanvasObject // the ranch's upgrades, the breeds and the expeditions
	expeditionsBody fyne.CanvasObject
	noTrips         *widget.Label // says that no trip is open to the monster, when none is
	trips           []tripRow
}

// tripRow is an expedition in the list: where, and when.
type tripRow struct {
	name, when *widget.Label
	object     fyne.CanvasObject
}

// facilityRow is a facility of the ranch: its name, how far it is as dots, what it gives, its next step and, in words,
// where that step stands.
type facilityRow struct {
	name, gives, next, offer *widget.Label
	pips                     *Pips
	state                    *ToneBox
}

func NewDashboard(svc profile.ProfileService) *Dashboard {
	d := &Dashboard{svc: svc}
	label := func() *widget.Label { return widget.NewLabel("") }
	box := func() *ToneBox { return NewToneBox() }

	d.gutsPerSecond = widget.NewCheck("per sec", func(on bool) { d.opts.GutsPerSecond = on; d.redraw() })
	d.growthFromZero = widget.NewCheck("from 0", func(on bool) { d.opts.GrowthFromZero = on; d.redraw() })
	d.status = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	d.status.Hide()

	d.breed = NewToneBox().Large()
	d.givenName, d.genus, d.age, d.stage, d.lifespan, d.money = label(), label(), label(), label(), label(), label()
	d.lifespanBar, d.lifeIndexBar, d.lifeIndex = NewMeter(), NewMeter(), box()

	for stat := range d.stats {
		d.stats[stat], d.statBars[stat], d.growth[stat] = label(), NewMeter().Colored(StatColors[stat]), box()
		d.statNames[stat] = NewColorText(Segment{Text: StatNames[stat], Color: StatColors[stat]})
	}
	d.statTotal, d.fatigue, d.stress = label(), label(), label()
	d.fatigueBar, d.stressBar, d.natureGauge, d.nature = NewMeter(), NewMeter(), NewGauge(), box()
	d.loyalty, d.rank, d.form, d.natureBase, d.lifeType, d.playType = label(), label(), label(), label(), label(), label()
	d.prize, d.nextSale, d.drug, d.drugTip = label(), label(), label(), label()
	d.tournament = label()
	d.tournamentRow = fact("IMA-FIMBA", d.tournament)
	d.tournamentRow.Hide()
	d.drugTip.Wrapping = fyne.TextWrapWord
	d.drugTip.Hide()
	d.fame, d.guts, d.arena, d.specials, d.like, d.dislike = box(), box(), box(), box(), box(), box()
	d.goldPeach, d.silverPeach = NewFlag("Gold Peach", ToneOrange), NewFlag("Silver Peach", ToneGray)
	d.itemUsed, d.errantry = NewFlag("Item Used", ToneLightGreen), NewFlag("Err. Sale Ready", ToneLightGreen)
	d.cupJellies, d.cocoonDone = box(), NewFlag("Cocoon Ready", ToneLimeGreen)
	d.date, d.fatigueWord, d.record, d.history = label(), Quiet(""), label(), label()
	d.date.Hide()
	d.ranchBody = d.buildRanch()
	d.ranchBody.Hide()
	d.expeditionsBody = d.buildExpeditions()
	d.buildBreeds()
	d.ranchBox = container.NewVBox(d.ranchBody, d.breeds, d.expeditionsBody)

	d.headerBody = d.buildHeader()
	d.vitalsBody = d.buildVitals()
	d.header = container.NewVBox(d.status, d.headerBody)
	d.vitals = container.NewVScroll(d.vitalsBody)
	return d
}

// barRow is a bar between its name and what it says, in the columns that every such row shares.
func barRow(name string, bar fyne.CanvasObject, rest ...fyne.CanvasObject) fyne.CanvasObject {
	return namedBarRow(widget.NewLabel(name), bar, rest...)
}

// namedBarRow is a barRow of a name that is not just text, such as a stat's in its colour.
func namedBarRow(name, bar fyne.CanvasObject, rest ...fyne.CanvasObject) fyne.CanvasObject {
	cells := append([]fyne.CanvasObject{name, bar}, rest...)
	return container.New(Columns{nameWidth, 0, valueWidth, chipWidth}, cells...)
}

// Quiet is a label in the quiet colour, for the names of the things shown, so that their values stand out.
func Quiet(text string) *widget.Label {
	l := widget.NewLabel(text)
	l.Importance = widget.LowImportance
	return l
}

// fact is a name and its value side by side; a value that is a coloured box stays the size of its text.
func fact(name string, value fyne.CanvasObject) fyne.CanvasObject {
	return container.NewHBox(Quiet(name), value)
}

func (d *Dashboard) buildHeader() fyne.CanvasObject {
	who := container.New(&d.whoRow,
		container.NewHBox(d.breed, d.givenName, d.genus),
		container.NewHBox(d.age, d.stage, d.date, widget.NewLabel("Money"), d.money),
	)
	life := container.New(&d.lifeRow,
		container.New(Columns{80, 0, 120}, widget.NewLabel("Lifespan"), d.lifespanBar, d.lifespan),
		container.New(Columns{80, 0, 140}, widget.NewLabel("Life Index"), d.lifeIndexBar, d.lifeIndex),
	)
	d.lifeRow.Even = true
	// The flags are under the name, where they are seen at once: they need more room than the header has beside it.
	flags := container.NewHBox(d.goldPeach, d.silverPeach, d.itemUsed, d.errantry)
	return container.NewVBox(who, flags, life)
}

// Compact puts each row of the header on two lines, who it is over how old, and the life span over the life index, to
// make the header narrow enough for a window of one column; or back on one line each.
func (d *Dashboard) Compact(on bool) {
	d.whoRow.Stacked, d.lifeRow.Stacked = on, on
	d.headerBody.Refresh()
}

// buildVitals puts what a breeder reads first on top, the stats and then how worn the monster is, and what is
// seldom needed at the bottom, as the facts that are either so or not.
func (d *Dashboard) buildVitals() fyne.CanvasObject {
	stats := container.NewVBox(
		container.New(Columns{nameWidth, 0, valueWidth, chipWidth}, Quiet("Stats"), layout.NewSpacer(), layout.NewSpacer(), d.growthFromZero),
	)
	for stat := range StatNames {
		stats.Add(namedBarRow(d.statNames[stat], d.statBars[stat], d.stats[stat], d.growth[stat]))
	}
	stats.Add(barRow("BST", layout.NewSpacer(), d.statTotal))

	condition := container.NewVBox(
		barRow("Fatigue", d.fatigueBar, d.fatigue, d.fatigueWord),
		barRow("Stress", d.stressBar, d.stress),
		barRow("Nature", d.natureGauge, d.nature),
	)

	pair := func(left, right fyne.CanvasObject) fyne.CanvasObject {
		return container.NewGridWithColumns(2, left, right)
	}
	facts := container.NewVBox(
		pair(fact("Loyalty", d.loyalty), fact("Rank", d.rank)),
		pair(fact("Fame", d.fame), fact("Form", d.form)),
		pair(fact("Guts Rate", container.NewHBox(d.guts, d.gutsPerSecond)), fact("Arena SPD", d.arena)),
		pair(fact("Base Nature", d.natureBase), fact("Play Type", d.playType)),
		pair(fact("Desired", d.like), fact("Undesired", d.dislike)),
		pair(fact("Prize Money", d.prize), fact("Shop Sale in", d.nextSale)),
		d.tournamentRow,
		fact("Record", d.record),
		fact("History", d.history),
		fact("Aging Type", d.lifeType),
		fact("Specials", d.specials),
		fact("Drug", d.drug),
		d.drugTip,
	)

	info := widget.NewLabel(cocoonInfo)
	info.Wrapping = fyne.TextWrapWord
	d.cocoon = container.NewVBox(
		widget.NewSeparator(),
		container.NewHBox(Quiet("Cup Jellies Eaten"), d.cupJellies, d.cocoonDone),
		widget.NewAccordion(widget.NewAccordionItem("Cocooning information", info)),
	)
	d.cocoon.Hide()

	return container.NewVBox(
		stats,
		widget.NewSeparator(),
		condition,
		widget.NewSeparator(),
		facts,
		d.cocoon,
	)
}

// section is the heading of a part of the page: its title, and to the right of it a line that says when.
func section(title string, when *widget.Label) fyne.CanvasObject {
	return container.NewHBox(widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), layout.NewSpacer(), when)
}

// buildRanch is the ranch's upgrades, under a heading that says when the game offers them: for the house, the
// stable and the altar, how far each is as dots, what it gives, its next step and what it costs, and in words where that
// step stands.
func (d *Dashboard) buildRanch() fyne.CanvasObject {
	d.upgradeDay = Quiet("")
	rows := container.NewVBox(section("Upgrades", d.upgradeDay))
	for i := range d.facilities {
		row := &d.facilities[i]
		row.name = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
		row.gives, row.next, row.offer = Quiet(""), Quiet(""), Quiet("")
		row.pips, row.state = NewPips(), NewToneBox()
		rows.Add(container.NewVBox(
			container.NewHBox(row.name, row.pips, row.gives, layout.NewSpacer(), row.next),
			container.NewHBox(row.state, layout.NewSpacer()), // as wide as its words, under the name
			row.offer,
		))
	}
	return rows
}

// buildBreeds is the main breeds, lit where unlocked, under a title with how many are.
func (d *Dashboard) buildBreeds() {
	flags := make([]fyne.CanvasObject, player.BreedCount)
	for i := range d.breedFlags {
		d.breedFlags[i] = NewFlag("", ToneLimeGreen)
		flags[i] = d.breedFlags[i]
	}
	d.breedsTitle = widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	d.breeds = container.NewVBox(widget.NewSeparator(), d.breedsTitle, container.NewGridWithColumns(5, flags...))
	d.breeds.Hide()
}

// buildExpeditions is the list of expeditions that the monster is fit for, nearest first: for each trip its name and
// when it comes.
func (d *Dashboard) buildExpeditions() fyne.CanvasObject {
	d.noTrips = Quiet("")
	box := container.NewVBox(widget.NewSeparator(), section("Expeditions", Quiet("")), d.noTrips)
	for range domain.Expeditions {
		row := &tripRow{
			name: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			when: widget.NewLabel(""),
		}
		row.object = container.NewHBox(row.name, layout.NewSpacer(), row.when)
		d.trips = append(d.trips, *row)
		box.Add(row.object)
	}
	box.Hide()
	return box
}

// Ranch is the panel under the training: the ranch's upgrades, which are there when the game keeps a ranch to
// read, and the expeditions.
func (d *Dashboard) Ranch() fyne.CanvasObject { return d.ranchBox }

// Header is the top of the dashboard: who the monster is, its age, its lifespan and its life index.
func (d *Dashboard) Header() fyne.CanvasObject { return d.header }

// Vitals is the body of the monster: its stats, its condition and everything else the game keeps of it.
func (d *Dashboard) Vitals() fyne.CanvasObject { return d.vitals }

// Load asks the service for the current profile and shows it. It is safe to call from any goroutine.
func (d *Dashboard) Load(ctx context.Context) {
	p, err := d.svc.Current(ctx)
	fyne.Do(func() {
		if d.seen.Changed(p, err) {
			d.apply(p, err)
		}
	})
}

// apply shows a profile, or why there is none. It must run on the UI thread.
func (d *Dashboard) apply(p domain.Profile, err error) {
	if err != nil {
		d.last = nil
		d.status.SetText(status.Message(err))
		d.status.Show()
		d.headerBody.Hide()
		d.vitalsBody.Hide()
		d.ranchBody.Hide()
		d.expeditionsBody.Hide()
		return
	}
	d.last = &p
	d.status.Hide()
	d.headerBody.Show()
	d.vitalsBody.Show()
	d.expeditionsBody.Show()
	d.show(NewMonsterView(p, d.opts))
}

// redraw shows the last profile again, for an option that changed.
func (d *Dashboard) redraw() {
	if d.last != nil {
		d.show(NewMonsterView(*d.last, d.opts))
	}
}

func (d *Dashboard) show(v MonsterView) {
	d.breed.Set(v.Breed)
	d.givenName.SetText(v.GivenName)
	d.genus.SetText(v.MainGenus + " / " + v.SubGenus)
	d.age.SetText(v.Age)
	d.stage.SetText(v.LifeStage)
	d.lifespan.SetText(v.Lifespan + " / " + v.InitialLifespan)
	d.lifespanBar.Set(v.Bars.Lifespan.Ratio, v.Bars.Lifespan.Tone)
	d.lifeIndex.Set(v.LifeIndex)
	d.lifeIndexBar.Set(v.Bars.LifeIndex.Ratio, v.Bars.LifeIndex.Tone)
	d.money.SetText(v.Money)

	for stat := range d.stats {
		d.stats[stat].SetText(v.Stats[stat])
		d.statBars[stat].Set(v.Bars.Stats[stat], ToneNone)
		d.growth[stat].Set(v.Growth[stat])
	}
	d.statTotal.SetText(v.StatTotal)
	d.fatigue.SetText(v.Fatigue)
	d.fatigueBar.Set(v.Bars.Fatigue.Ratio, v.Bars.Fatigue.Tone)
	d.stress.SetText(v.Stress)
	d.stressBar.Set(v.Bars.Stress.Ratio, v.Bars.Stress.Tone)
	d.nature.Set(v.Nature)
	d.natureGauge.Set(v.Bars.Nature.Ratio, ToneNone)

	d.loyalty.SetText(v.Loyalty + " (spoil " + v.Spoil + " / fear " + v.Fear + ")")
	d.rank.SetText(v.Rank)
	d.fame.Set(v.Fame)
	d.form.SetText(v.Form)
	d.natureBase.SetText(v.NatureBase)
	d.guts.Set(v.GutsRate)
	d.arena.Set(v.ArenaSpeed)
	d.specials.Set(v.Specials)
	d.like.Set(v.Like)
	d.dislike.Set(v.Dislike)
	d.lifeType.SetText(v.LifeType)
	d.playType.SetText(v.PlayType)
	d.prize.SetText(v.PrizeMoney)
	d.nextSale.SetText(v.NextSale)
	d.drug.SetText(strings.TrimSpace(v.Drug + " " + v.DrugWeeks))
	d.drugTip.SetText(v.DrugTip)
	setVisible(d.drugTip, v.DrugTip != "")
	d.goldPeach.Set(v.GoldPeach)
	d.silverPeach.Set(v.SilverPeach)
	d.itemUsed.Set(v.ItemUsed)
	d.errantry.Set(v.ErrantrySale)

	setVisible(d.cocoon, v.Cocoon != nil)
	if v.Cocoon != nil {
		d.cupJellies.Set(v.Cocoon.CupJellies)
		d.cocoonDone.Set(v.Cocoon.Ready)
	}

	d.fatigueWord.SetText(v.FatigueWord)
	d.record.SetText(v.Record)
	d.history.SetText(v.History)
	d.showRanch(v.Ranch)
	d.showExpeditions(v.Expeditions)
}

// showExpeditions shows the trips that the monster is fit for, nearest first, or says that there are none.
func (d *Dashboard) showExpeditions(e ExpeditionsView) {
	d.noTrips.SetText(e.None)
	setVisible(d.noTrips, e.None != "")
	for i, row := range d.trips {
		if i >= len(e.Trips) {
			row.object.Hide()
			continue
		}
		trip := e.Trips[i]
		row.object.Show()
		row.name.SetText(trip.Name)
		row.when.SetText(trip.When)
	}
}

// showRanch shows the ranch's date and its panel of upgrades, or hides them when the game keeps no ranch.
func (d *Dashboard) showRanch(r RanchView) {
	setVisible(d.date, r.Known)
	setVisible(d.ranchBody, r.Known)
	d.date.SetText(r.Date)
	d.tournament.SetText(r.Tournament)
	setVisible(d.tournamentRow, r.Tournament != "")
	d.upgradeDay.SetText("offered " + r.UpgradeDay)
	for i, f := range r.Facilities {
		row := d.facilities[i]
		row.name.SetText(f.Name)
		row.pips.Set(f.Level, f.Max)
		row.gives.SetText(f.Gives)
		setVisible(row.gives, f.Gives != "")
		row.next.SetText(f.Next)
		setVisible(row.next, f.Next != "")
		row.offer.SetText(f.Offer)
		setVisible(row.offer, f.Offer != "")
		row.state.Set(f.State)
	}
	for i, flag := range d.breedFlags {
		if i >= len(r.Breeds) {
			flag.Hide()
			continue
		}
		flag.Show()
		flag.SetText(r.Breeds[i].Name)
		flag.Set(r.Breeds[i].Unlocked)
	}
	d.breedsTitle.SetText(r.BreedsTitle)
	setVisible(d.breeds, len(r.Breeds) > 0)
}
