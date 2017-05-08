package parsers

var dscanTestCases = []Case{
	{
		"Simple",
		`+	Noctis	3,225 m
+	Thrasher	12 km
some dude's Stabber Fleet Issue	Stabber Fleet Issue	-
Wreck	Tayra	82 km`,
		&DScan{
			Items: []DScanItem{
				DScanItem{Name: "Noctis", Distance: 3225, DistanceUnit: "m"},
				DScanItem{Name: "Stabber Fleet Issue", Distance: 0, DistanceUnit: ""},
				DScanItem{Name: "Tayra", Distance: 82, DistanceUnit: "km"},
				DScanItem{Name: "Thrasher", Distance: 12, DistanceUnit: "km"},
			},
			lines: []int{0, 1, 2, 3},
		},
		Input{},
		true,
	}, {
		"non-breakable space in distance",
		"test	Noctis	3\xc2\xa0225 m",
		&DScan{
			Items: []DScanItem{DScanItem{Name: "Noctis", Distance: 3225, DistanceUnit: "m"}},
			lines: []int{0},
		},
		Input{},
		true,
	}, {
		"that's no moon!",
		"Otanuomi V - Moon 11	Moon	10.4 AU",
		&DScan{
			Items: []DScanItem{DScanItem{Name: "Moon", Distance: 104, DistanceUnit: "AU"}},
			lines: []int{0},
		},
		Input{},
		true,
	},
}
