package domain

func IsDrug(id int) bool { return id > 33 && id < 42 }

var drugTips = map[int]string{
	34: "Teromeann:\nCombat stimulant. Increases SPD by 100%, and POW by 50% in battle for one week.\n-20 Weeks of lifespan on use.",
	37: "Kasseitan:\nCombat stimulant. Increases SPD by 50%, and POW by 50% in battle for one week.\n-20 Weeks of lifespan on use.",
	38: "Troron:\nTraining stimulant. Increases POW by 10 and SKI by 5 for each Success/Cheat/Great for four weeks.\n-6 Weeks of lifespan on use.",
	39: "Nageel:\nCombat stimulant. Increases SKI by 100%, and DEF by 50% in battle for one week.\n-20 Weeks of lifespan on use.",
	40: "Torokachin FX:\nTraining stimulant. Increases POW and SKI by 25 for each Success/Cheat/Great for four weeks.\n-12 Weeks of lifespan on use. (Also; the user has edited this drug in.)",
	41: "Paradoxine:\nTraining stimulant. Increases POW and SKI by 30 for each Success/Cheat/Great for four weeks.\nEach increase also decreases SPD and DEF by 10%.\n-18 Weeks of lifespan on use.",
}

func DrugTip(id int) string { return drugTips[id] }
