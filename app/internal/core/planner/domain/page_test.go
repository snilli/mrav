package domain_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	monster "app/internal/core/monster/domain"
	"app/internal/core/planner/domain"
)

type pageRow struct {
	ID   string `json:"id"`
	Item string `json:"item"`
	Act  string `json:"act"`
	Food string `json:"food"`
	Ftg  string `json:"ftg"`
	Str  string `json:"str"`
	Li   string `json:"li"`
	Wl   string `json:"wl"`
}

type pageCase struct {
	Name     string `json:"name"`
	Settings struct {
		Version    string `json:"version"`
		Main       int    `json:"main"`
		Sub        int    `json:"sub"`
		Like       string `json:"like"`
		Dislike    string `json:"dislike"`
		LifeStage  int    `json:"life_stage"`
		Form       int    `json:"form"`
		Pots       int    `json:"pots"`
		Sculptures int    `json:"sculptures"`
		Fatigue    int    `json:"fatigue"`
		Stress     int    `json:"stress"`
	} `json:"settings"`
	Steps []struct {
		Item   string `json:"item"`
		Action string `json:"action"`
	} `json:"steps"`
	Foods   map[string]string `json:"foods"`
	Popups  []string          `json:"popups"`
	Rows    []pageRow         `json:"rows"`
	Lived   int               `json:"lived"`
	Weeks   int               `json:"weeks"`
	Average float64           `json:"average"`
}

func readJSON(name string, into any) {
	raw, err := os.ReadFile("testdata/" + name)
	Expect(err).NotTo(HaveOccurred())
	Expect(json.Unmarshal(raw, into)).To(Succeed())
}

func foodOf(key string) domain.Food {
	if key == "none" {
		return domain.NoFood
	}
	food, ok := domain.ParseFood(key)
	Expect(ok).To(BeTrue(), key)
	return food
}

var pageItemNames = map[string]string{"MB (+10 Loyalty)": "MB +10", "MB (+0 Loyalty)": "MB +0", "MB (-10 Loyalty)": "MB -10"}

func itemOf(text string) domain.Item {
	if text == "" {
		return domain.NoItem
	}
	if name, ok := pageItemNames[text]; ok {
		text = name
	}
	for _, it := range domain.Items() {
		if it.Name() == text {
			return it
		}
	}
	Fail("no item called " + text)
	return domain.NoItem
}

func actionOf(text string) domain.Action {
	switch text {
	case "":
		return domain.NoAction
	case "-":
		return domain.Away
	}
	for _, a := range domain.Actions() {
		if a.Name() == text {
			return a
		}
	}
	Fail("no action called " + text)
	return domain.NoAction
}

func itemText(it domain.Item) string {
	if it == domain.NoItem {
		return ""
	}
	for page, name := range pageItemNames {
		if name == it.Name() {
			return page
		}
	}
	return it.Name()
}

func actionText(a domain.Action) string {
	switch a {
	case domain.NoAction:
		return ""
	case domain.Away:
		return "-"
	}
	return a.Name()
}

func shown(r domain.Row) pageRow {
	out := pageRow{ID: fmt.Sprintf("%d-%d", r.Month, r.Week)}
	if r.Kind == domain.FoodRow {
		out.Food = r.Food.Key()
	} else {
		out.Item, out.Act = itemText(r.Item), actionText(r.Action)
	}
	if r.Shown {
		out.Ftg, out.Str = strconv.Itoa(r.Fatigue), strconv.Itoa(r.Stress)
	}
	if r.ShowLI {
		out.Li = strconv.Itoa(r.LI)
	}
	if r.ShowLost {
		out.Wl = strconv.Itoa(r.Lost)
		if r.Lost > 0 {
			out.Wl = "+" + out.Wl
		}
	}
	return out
}

func (c pageCase) settings() domain.Settings {
	version := monster.MR2
	if c.Settings.Version == "mr2dx" {
		version = monster.MR2DX
	}
	return domain.Settings{
		Version: version, Main: c.Settings.Main, Sub: c.Settings.Sub, Like: foodOf(c.Settings.Like), Dislike: foodOf(c.Settings.Dislike),
		LifeStage: c.Settings.LifeStage, Form: c.Settings.Form, Pots: c.Settings.Pots, Sculptures: c.Settings.Sculptures,
		Fatigue: c.Settings.Fatigue, Stress: c.Settings.Stress,
	}
}

func (c pageCase) steps() []domain.Step {
	var out []domain.Step
	for _, s := range c.Steps {
		out = append(out, domain.Step{Item: itemOf(s.Item), Action: actionOf(s.Action)})
	}
	return out
}

func (c pageCase) foods() map[int]domain.Food {
	out := map[int]domain.Food{}
	for month, key := range c.Foods {
		m, err := strconv.Atoi(month)
		Expect(err).NotTo(HaveOccurred())
		out[m] = foodOf(key)
	}
	return out
}

func (c pageCase) refusedAnExpedition() bool {
	for _, popup := range c.Popups {
		if strings.Contains(popup, "Expedition has to start") {
			return true
		}
	}
	return false
}

func shippedReactions() domain.Reactions {
	var rows []struct {
		Genus    int    `json:"genus"`
		Food     string `json:"food"`
		Stage    string `json:"stage"`
		Reaction string `json:"reaction"`
	}
	readJSON("reactions.json", &rows)
	var list []domain.FoodReaction
	for _, r := range rows {
		stage, ok := domain.ParseStage(r.Stage)
		Expect(ok).To(BeTrue(), r.Stage)
		reaction, ok := domain.ParseReaction(r.Reaction)
		Expect(ok).To(BeTrue(), r.Reaction)
		list = append(list, domain.FoodReaction{Genus: r.Genus, Food: foodOf(r.Food), Stage: stage, Reaction: reaction})
	}
	return domain.NewReactions(list)
}

var _ = Describe("a plan worked out the way the Training Planner of LegendCup works it out", func() {
	var cases []pageCase
	readJSON("page.json", &cases)

	It("has the cases that the page worked out, with every action, every item and both games", func() {
		Expect(len(cases)).To(BeNumerically(">=", 270))
	})

	for _, c := range cases {
		if c.refusedAnExpedition() {
			continue
		}
		It(c.Name, func() {
			plan := domain.Build(shippedReactions(), shippedEffects(), c.settings(), c.steps(), c.foods())

			var got []pageRow
			for _, r := range plan.Rows {
				got = append(got, shown(r))
			}
			Expect(got).To(Equal(c.Rows), "the rows")
			Expect(plan.Totals.Lived).To(Equal(c.Lived), "weeks lived")
			Expect(plan.Totals.Weeks).To(Equal(c.Weeks), "weeks lost or gained")
			Expect(plan.Totals.Average).To(BeNumerically("~", c.Average, 1e-9), "average a month")
			Expect(plan.Problems).To(HaveLen(len(c.Popups)), "what the page refused")
		})
	}
})
