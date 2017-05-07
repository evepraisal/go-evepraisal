package parsers

var killmailTestCases = []Case{
	{
		"Basic",
		`2013.06.15 17:28:00

Victim: Some poor victim
Corp: Victim's Corp Name
Alliance: Victim's Alliance Name
Faction: Unknown
Destroyed: Scimitar
System: Jita
Security: 0.9
Damage Taken: 14194

Involved parties:

Name: Ganker Name (laid the final blow)
Security: -1.00
Corp: Ganker Corp
Alliance: Ganker Alliance
Faction: Unknown
Ship: Apocalypse Navy Issue
Weapon: Mega Pulse Laser II
Damage Done: 14194

Name: Ganker Name2
Security: -10.00
Corp: Ganker Corp
Alliance: Ganker Alliance
Faction: Unknown
Ship: Rifter
Weapon: Some tiny little gun
Damage Done: 0

Destroyed items:

Medium Armor Maintenance Bot I, Qty: 3 (Drone Bay)
Tengu Engineering - Capacitor Regeneration Matrix
Power Diagnostic System II (Cargo)

Dropped items:

Warrior II (Drone Bay)`,
		&Killmail{
			datetime: "2013.06.15 17:28:00",
			victim: map[string]interface{}{
				"faction":      "Unknown",
				"destroyed":    "Scimitar",
				"system":       "Jita",
				"security":     "0.9",
				"damage_taken": "14194",
				"victim":       "Some poor victim",
				"corp":         "Victim's Corp Name",
				"alliance":     "Victim's Alliance Name"},
			involved: []map[string]interface{}{
				map[string]interface{}{
					"killing_blow": true,
					"corp":         "Ganker Corp",
					"alliance":     "Ganker Alliance",
					"faction":      "Unknown",
					"ship":         "Apocalypse Navy Issue",
					"name":         "Ganker Name",
					"security":     "-1.00",
					"weapon":       "Mega Pulse Laser II",
					"damage_done":  "14194"},
				map[string]interface{}{
					"weapon":      "Some tiny little gun",
					"damage_done": "0",
					"name":        "Ganker Name2",
					"security":    "-10.00",
					"corp":        "Ganker Corp",
					"alliance":    "Ganker Alliance",
					"faction":     "Unknown",
					"ship":        "Rifter"}},
			destroyed: []KillmailItem{
				KillmailItem{name: "Medium Armor Maintenance Bot I", quantity: 3, location: "Drone Bay"},
				KillmailItem{name: "Tengu Engineering - Capacitor Regeneration Matrix", quantity: 1, location: ""},
				KillmailItem{name: "Power Diagnostic System II", quantity: 1, location: "Cargo"},
			},
			dropped:   []KillmailItem{KillmailItem{name: "Warrior II", quantity: 1, location: "Drone Bay"}},
			lineCount: 40,
		},
		nil,
		true,
	}, {
		"Empty String", ``, nil, Input{0: ""}, true,
	}, {
		"Truncated",
		`2013.06.15 17:28:00
Victim: Some poor victim
Corp: Victim's Corp Name
Damage Taken`,
		nil,
		Input{0: "2013.06.15 17:28:00", 1: "Victim: Some poor victim", 2: "Corp: Victim's Corp Name", 3: "Damage Taken"},
		false,
	}, {
		"Short",
		`2013.06.15 17:28:00

Victim: Some poor victim`,
		&Killmail{
			datetime:  "2013.06.15 17:28:00",
			victim:    map[string]interface{}{"victim": "Some poor victim"},
			lineCount: 3,
		},
		nil,
		false,
	},
}

// 		KILLMAIL_TABLE.add_test(
//     '''2013.06.15 17:28:00
// Victim: Some poor victim
// Corp: Victim's Corp Name
// Damage Taken''', Unparsable)
