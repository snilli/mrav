package profileui

import (
	"context"
	"fmt"
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
	svc  profile.ProfileService
	opts Options
	last *domain.Profile // what is shown, to show it again when an option changes

	gutsPerSecond  *widget.Check // beside the guts rate, which it changes
	growthFromZero *widget.Check // above the growth rates, which it changes
	status         *widget.Label
	headerBody     fyne.CanvasObject
	vitalsBody     fyne.CanvasObject
	header         fyne.CanvasObject
	vitals         fyne.CanvasObject

	breed                                         *ToneBox
	givenName, genus, age, stage, lifespan, money *widget.Label
	lifespanBar, lifeIndexBar                     *Meter
	lifeIndex                                     *ToneBox

	stats                                         [monster.StatCount]*widget.Label
	statBars                                      [monster.StatCount]*Meter
	growth                                        [monster.StatCount]*ToneBox
	statTotal                                     *widget.Label
	fatigue, stress                               *widget.Label
	fatigueBar, stressBar, natureGauge            *Meter
	nature                                        *ToneBox
	loyalty, form, natureBase, lifeType, playType *widget.Label
	prize, nextSale, drug, drugTip                *widget.Label
	fame, guts, arena, specials, like, dislike    *ToneBox
	goldPeach, silverPeach, itemUsed, errantry    *Flag
	cocoon                                        fyne.CanvasObject
	cupJellies                                    *ToneBox
	cocoonDone                                    *Flag

	date                         *widget.Label // the game's date in the header, where the game keeps one
	fatigueWord, record, history *widget.Label
	ranchBody                    fyne.CanvasObject // the panel of upgrades, hidden where the game keeps no ranch
	ranchPanel                   *widget.Accordion // closed until opened, titled with the day of the upgrades
	ranchItem                    *widget.AccordionItem
	facilities                   [3]facilityRow
	breeds                       *widget.Accordion // the main breeds, lit where unlocked, opening from the count
	breedsItem                   *widget.AccordionItem
	breedFlags                   [player.BreedCount]*Flag

	ranchBox                       fyne.CanvasObject // the ranch's panel, and the expeditions under it
	expeditionsBody                fyne.CanvasObject
	expeditions                    *widget.Accordion // the trips, opening from the nearest one
	expeditionsItem                *widget.AccordionItem
	startEnergy, gameEnergy, costs *widget.Label
	trips                          []tripRow
}

// tripRow is an expedition in the list: where, when, and the three things its invitation asks of the monster.
type tripRow struct {
	name, when, note *widget.Label
	chips            [3]*ToneBox
	object           fyne.CanvasObject
}

// facilityRow is a facility of the ranch in the panel: its name, how far it is, and its next step.
type facilityRow struct {
	name, level, next *widget.Label
	bar               *Meter
	state             *ToneBox
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
		d.stats[stat], d.statBars[stat], d.growth[stat] = label(), NewMeter(), box()
	}
	d.statTotal, d.fatigue, d.stress = label(), label(), label()
	d.fatigueBar, d.stressBar, d.natureGauge, d.nature = NewMeter(), NewMeter(), NewGauge(), box()
	d.loyalty, d.form, d.natureBase, d.lifeType, d.playType = label(), label(), label(), label(), label()
	d.prize, d.nextSale, d.drug, d.drugTip = label(), label(), label(), label()
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
	cells := append([]fyne.CanvasObject{widget.NewLabel(name), bar}, rest...)
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
	who := container.NewHBox(
		d.breed, d.givenName, d.genus, layout.NewSpacer(),
		d.age, d.stage, d.date, widget.NewLabel("Money"), d.money,
	)
	life := container.NewGridWithColumns(2,
		container.New(Columns{80, 0, 120}, widget.NewLabel("Lifespan"), d.lifespanBar, d.lifespan),
		container.New(Columns{80, 0, 140}, widget.NewLabel("Life Index"), d.lifeIndexBar, d.lifeIndex),
	)
	return container.NewVBox(who, life)
}

// buildVitals puts what a breeder reads first on top, the stats and then how worn the monster is, and what is
// seldom needed at the bottom, as the facts that are either so or not.
func (d *Dashboard) buildVitals() fyne.CanvasObject {
	stats := container.NewVBox(
		container.New(Columns{nameWidth, 0, valueWidth, chipWidth}, Quiet("Stats"), layout.NewSpacer(), layout.NewSpacer(), d.growthFromZero),
	)
	for stat, name := range StatNames {
		stats.Add(barRow(name, d.statBars[stat], d.stats[stat], d.growth[stat]))
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
		fact("Loyalty", d.loyalty),
		pair(fact("Fame", d.fame), fact("Form", d.form)),
		pair(fact("Guts Rate", container.NewHBox(d.guts, d.gutsPerSecond)), fact("Arena SPD", d.arena)),
		pair(fact("Base Nature", d.natureBase), fact("Play Type", d.playType)),
		pair(fact("Desired", d.like), fact("Undesired", d.dislike)),
		pair(fact("Prize Money", d.prize), fact("Shop Sale in", d.nextSale)),
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
		widget.NewSeparator(),
		container.NewHBox(d.goldPeach, d.silverPeach, d.itemUsed, d.errantry),
	)
}

// buildRanch is the panel of the ranch's upgrades, closed until it is opened and titled with the day they are
// offered: for the house, the stable and the altar, how far each is, the next step and what it costs, and whether
// that step can be had.
func (d *Dashboard) buildRanch() fyne.CanvasObject {
	rows := container.NewVBox()
	for i := range d.facilities {
		row := &d.facilities[i]
		row.name, row.level, row.next = widget.NewLabel(""), widget.NewLabel(""), Quiet("")
		row.bar, row.state = NewMeter(), NewToneBox()
		rows.Add(container.NewVBox(
			container.New(Columns{70, 0, 44, 190}, row.name, row.bar, row.level, row.state),
			row.next,
		))
	}
	d.ranchItem = widget.NewAccordionItem("Ranch", rows)
	d.ranchPanel = widget.NewAccordion(d.ranchItem)
	return d.ranchPanel
}

// buildBreeds is the panel of the main breeds, lit where unlocked, closed until it is opened and titled with how
// many are.
func (d *Dashboard) buildBreeds() {
	flags := make([]fyne.CanvasObject, player.BreedCount)
	for i := range d.breedFlags {
		d.breedFlags[i] = NewFlag("", ToneLimeGreen)
		flags[i] = d.breedFlags[i]
	}
	d.breedsItem = widget.NewAccordionItem("Breeds unlocked", container.NewGridWithColumns(5, flags...))
	d.breeds = widget.NewAccordion(d.breedsItem)
	d.breeds.Hide()
}

// buildExpeditions is the list of expeditions, which opens from the nearest one: the energy a monster would start
// with, and for each trip when it comes and what its invitation asks of the monster.
func (d *Dashboard) buildExpeditions() fyne.CanvasObject {
	d.startEnergy, d.gameEnergy, d.costs = Quiet(""), Quiet(""), Quiet("")
	d.gameEnergy.Hide()
	d.costs.Wrapping = fyne.TextWrapWord
	body := container.NewVBox(d.startEnergy, d.gameEnergy)
	for range domain.Expeditions {
		row := &tripRow{
			name: widget.NewLabelWithStyle("", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
			when: widget.NewLabel(""),
			note: Quiet(""),
		}
		for i := range row.chips {
			row.chips[i] = NewToneBox()
		}
		row.object = container.NewVBox(
			container.NewHBox(row.name, layout.NewSpacer(), row.when),
			row.note,
			container.NewGridWithColumns(len(row.chips), row.chips[0], row.chips[1], row.chips[2]),
		)
		d.trips = append(d.trips, *row)
		body.Add(row.object)
	}
	body.Add(d.costs)
	d.expeditionsItem = widget.NewAccordionItem("Expeditions", body)
	d.expeditions = widget.NewAccordion(d.expeditionsItem)
	box := container.NewVBox(widget.NewSeparator(), d.expeditions)
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
	fyne.Do(func() { d.apply(p, err) })
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

// showExpeditions shows the trips, nearest first, and what the monster has for each.
func (d *Dashboard) showExpeditions(e ExpeditionsView) {
	d.expeditionsItem.Title = e.Title
	d.startEnergy.SetText(e.StartEnergy)
	d.gameEnergy.SetText(e.GameEnergy)
	setVisible(d.gameEnergy, e.GameEnergy != "")
	d.costs.SetText(e.Costs)
	for i, row := range d.trips {
		if i >= len(e.Trips) {
			row.object.Hide()
			continue
		}
		trip := e.Trips[i]
		row.object.Show()
		row.name.SetText(trip.Name)
		row.when.SetText(trip.When)
		row.note.SetText(trip.Note)
		setVisible(row.note, trip.Note != "")
		for j, chip := range row.chips {
			if j >= len(trip.Conditions) {
				continue
			}
			c := trip.Conditions[j]
			mark, tone := "✗", ToneOrange
			if c.Met {
				mark, tone = "✓", ToneLimeGreen
			}
			chip.Set(Box{Text: fmt.Sprintf("%s %s (%s)", mark, c.Text, c.Have), Tone: tone})
		}
	}
	d.expeditions.Refresh()
}

// showRanch shows the ranch's date and its panel of upgrades, or hides them when the game keeps no ranch.
func (d *Dashboard) showRanch(r RanchView) {
	setVisible(d.date, r.Known)
	setVisible(d.ranchBody, r.Known)
	d.date.SetText(r.Date)
	d.ranchItem.Title = "Ranch · House & stable: " + r.UpgradeDay
	d.ranchPanel.Refresh()
	for i, f := range r.Facilities {
		row := d.facilities[i]
		row.name.SetText(f.Name)
		row.level.SetText(fmt.Sprintf("%d/%d", f.Level, f.Max))
		row.bar.Set(f.Ratio, ToneNone)
		next := f.Next
		if f.Detail != "" {
			next = f.Detail + " · " + next
		}
		row.next.SetText(next)
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
	d.breedsItem.Title = r.BreedsTitle
	setVisible(d.breeds, len(r.Breeds) > 0)
	d.breeds.Refresh()
}
