package parsers

var dscanTestCases = []Case{
	{
		"Simple",
		`+	Noctis	3,225 m
+	Thrasher	12 km
some dude's Stabber Fleet Issue	Stabber Fleet Issue	-
Wreck	Tayra	82 km`,
		&DScan{
			items: []DScanItem{
				DScanItem{name: "Noctis", distance: 3225, distanceUnit: "m"},
				DScanItem{name: "Thrasher", distance: 12, distanceUnit: "km"},
				DScanItem{name: "Stabber Fleet Issue", distance: 0, distanceUnit: ""},
				DScanItem{name: "Tayra", distance: 82, distanceUnit: "km"},
			},
			raw: []string{"+\tNoctis\t3,225 m", "+\tThrasher\t12 km", "some dude's Stabber Fleet Issue\tStabber Fleet Issue\t-", "Wreck\tTayra\t82 km"},
		},
		nil,
		true,
	}, {
		"non-breakable space in distance",
		"test	Noctis	3\xc2\xa0225 m",
		&DScan{
			items: []DScanItem{DScanItem{name: "Noctis", distance: 3225, distanceUnit: "m"}},
			raw:   []string{"test\tNoctis\t3\u00a0225 m"},
		},
		nil,
		true,
	}, {
		"that's no moon!",
		"Otanuomi V - Moon 11	Moon	10.4 AU",
		&DScan{
			items: []DScanItem{DScanItem{name: "Moon", distance: 104, distanceUnit: "AU"}},
			raw:   []string{"Otanuomi V - Moon 11\tMoon\t10.4 AU"},
		},
		nil,
		true,
	},
}
