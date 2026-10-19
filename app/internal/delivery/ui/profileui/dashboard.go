package profileui

import (
	"context"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	monster "app/internal/core/monster/domain"
	"app/internal/core/profile"
	"app/internal/core/profile/domain"
	"app/internal/delivery/ui/status"
)

// StatNames are the stats as MR2AV abbreviates them, by stat.
var StatNames = [monster.StatCount]string{"LIF", "POW", "INT", "SKI", "SPD", "DEF"}

// The widths of the columns of a row of bars: the name, then the bar, then what it says, then a growth.
const (
	nameWidth  = 70
	valueWidth = 110
	chipWidth  = 64
)

// cocoonInfo is what MR2AV tells a breeder about cocooning.
const cocoonInfo = `Cocooning requires a Worm that has:
- Zero fatigue.
- Less than 30 Stress.
- *80+ Loyalty.
- C Rank or lower.
- Lived for at least 4 years, and is on the Ranch when the calendar reaches 'June, Week 4.'

If the 'Cup Jellies Eaten' counter is green, you have fed enough Cup Jellies for your Worm to become a Beaclon.
If the box is not green, but 'Cocoon Ready' is ticked, your Worm will cocoon into a random /Worm monster.

If this cocooning information disappears, your Worm is past C Rank, therefore it cannot cocoon any more.

*79 Loyalty is possible, with one of Spoil/Fear being 79 and the other 80.
78 Loyalty is possible with both Spoil and Fear being 79.
'Cocoon Ready' will only tick on 80+ Loyalty, despite these two cases.`

// Dashboard is the monster the game shows, on one screen: a header of who it is and how much life it has left, and
// the vitals of its body. It is two parts so that the window can put them where it likes, and one value so that
// they load together.
type Dashboard struct {
	svc  profile.ProfileService
	opts Options
	last *domain.Profile // what is shown, to show it again when an option changes

	gutsPerSecond  *widget.Check
	growthFromZero *widget.Check
	status         *widget.Label
	headerBody     fyne.CanvasObject
	vitalsBody     fyne.CanvasObject
	header         fyne.CanvasObject
	vitals         fyne.CanvasObject

	breed                                         *ToneBox
	givenName, genus, age, stage, lifespan, money *widget.Label
	lifespanBar, lifeIndexBar                     *Meter
	lifeIndex                                     *ToneBox

	stats                                                  [monster.StatCount]*widget.Label
	statBars                                               [monster.StatCount]*Meter
	growth                                                 [monster.StatCount]*ToneBox
	statTotal                                              *widget.Label
	fatigue, stress                                        *widget.Label
	fatigueBar, stressBar, natureGauge                     *Meter
	nature                                                 *ToneBox
	loyalty, form, natureBase, lifeType, playType          *widget.Label
	prize, nextSale, drug, drugTip                         *widget.Label
	fame, guts, arena, specials, like, dislike             *ToneBox
	goldPeach, silverPeach, itemUsed, errantry, cocoonDone *widget.Check
	cocoon                                                 fyne.CanvasObject
	cupJellies                                             *ToneBox
}

func NewDashboard(svc profile.ProfileService) *Dashboard {
	d := &Dashboard{svc: svc}
	label := func() *widget.Label { return widget.NewLabel("") }
	box := func() *ToneBox { return NewToneBox() }
	check := func(text string) *widget.Check {
		c := widget.NewCheck(text, nil)
		c.Disable() // they show state; the game, not the viewer, changes it
		return c
	}

	d.gutsPerSecond = widget.NewCheck("Guts per second", func(on bool) { d.opts.GutsPerSecond = on; d.redraw() })
	d.growthFromZero = widget.NewCheck("Growth numbers from 0", func(on bool) { d.opts.GrowthFromZero = on; d.redraw() })
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
	d.goldPeach, d.silverPeach = check("Gold Peach"), check("Silver Peach")
	d.itemUsed, d.errantry = check("Item Used"), check("Err. Sale Ready")
	d.cupJellies, d.cocoonDone = box(), check("Cocoon Ready")

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

func (d *Dashboard) buildHeader() fyne.CanvasObject {
	who := container.NewHBox(
		d.breed, d.givenName, d.genus, layout.NewSpacer(),
		d.age, d.stage, widget.NewLabel("Money"), d.money,
	)
	life := container.NewGridWithColumns(2,
		container.New(Columns{80, 0, 120}, widget.NewLabel("Lifespan"), d.lifespanBar, d.lifespan),
		container.New(Columns{80, 0, 140}, widget.NewLabel("Life Index"), d.lifeIndexBar, d.lifeIndex),
	)
	return container.NewVBox(who, life)
}

func (d *Dashboard) buildVitals() fyne.CanvasObject {
	bars := container.NewVBox()
	for stat, name := range StatNames {
		bars.Add(barRow(name, d.statBars[stat], d.stats[stat], d.growth[stat]))
	}
	bars.Add(barRow("BST", layout.NewSpacer(), d.statTotal))
	bars.Add(barRow("Fatigue", d.fatigueBar, d.fatigue))
	bars.Add(barRow("Stress", d.stressBar, d.stress))
	bars.Add(barRow("Nature", d.natureGauge, d.nature))

	facts := widget.NewForm(
		widget.NewFormItem("Loyalty", d.loyalty),
		widget.NewFormItem("Fame", d.fame),
		widget.NewFormItem("Form", d.form),
		widget.NewFormItem("Base Nature", d.natureBase),
		widget.NewFormItem("Guts Rate", d.guts),
		widget.NewFormItem("Arena SPD", d.arena),
		widget.NewFormItem("Battle Specials", d.specials),
		widget.NewFormItem("Desired Item", d.like),
		widget.NewFormItem("Undesired Item", d.dislike),
		widget.NewFormItem("Aging Type", d.lifeType),
		widget.NewFormItem("Play Type", d.playType),
		widget.NewFormItem("Prize Money", d.prize),
		widget.NewFormItem("Shop Sale in", d.nextSale),
		widget.NewFormItem("Drug", container.NewVBox(d.drug, d.drugTip)),
	)

	info := widget.NewLabel(cocoonInfo)
	info.Wrapping = fyne.TextWrapWord
	d.cocoon = container.NewVBox(
		widget.NewSeparator(),
		widget.NewForm(widget.NewFormItem("Cup Jellies Eaten", d.cupJellies)),
		d.cocoonDone,
		widget.NewAccordion(widget.NewAccordionItem("Cocooning information", info)),
	)
	d.cocoon.Hide()

	return container.NewVBox(
		bars,
		widget.NewSeparator(),
		facts,
		container.NewGridWithColumns(2, d.goldPeach, d.silverPeach, d.itemUsed, d.errantry),
		d.cocoon,
		widget.NewSeparator(),
		d.gutsPerSecond, d.growthFromZero,
	)
}

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
		return
	}
	d.last = &p
	d.status.Hide()
	d.headerBody.Show()
	d.vitalsBody.Show()
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
	d.goldPeach.SetChecked(v.GoldPeach)
	d.silverPeach.SetChecked(v.SilverPeach)
	d.itemUsed.SetChecked(v.ItemUsed)
	d.errantry.SetChecked(v.ErrantrySale)

	setVisible(d.cocoon, v.Cocoon != nil)
	if v.Cocoon != nil {
		d.cupJellies.Set(v.Cocoon.CupJellies)
		d.cocoonDone.SetChecked(v.Cocoon.Ready)
	}
}
