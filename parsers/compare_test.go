package parsers

var compareTestCases = []Case{
	{
		"Simple",
		`Small Armor Repairer I	Tech I	40 GJ	5 MW	5 tf	6.00 s	69 HP	Level 0
Small Armor Repairer II	Tech II	40 GJ	6 MW	6 tf	6.00 s	92 HP	Level 5
'Gorget' Small Armor Repairer I	Storyline	40 GJ	5 MW	4 tf	6.00 s	92 HP	Level 6
Republic Fleet Small Armor Repairer	Faction	36 GJ	5 MW	4 tf	6.00 s	83 HP	Level 7
Centii C-Type Small Armor Repairer	Deadspace	45 GJ	5 MW	4 tf	6.00 s	114 HP	Level 11`,
		&Compare{
			Items: []CompareItem{
				{Name: "'Gorget' Small Armor Repairer I"},
				{Name: "Centii C-Type Small Armor Repairer"},
				{Name: "Republic Fleet Small Armor Repairer"},
				{Name: "Small Armor Repairer II"},
				{Name: "Small Armor Repairer I"},
			},
			lines: []int{0, 1, 2, 3, 4},
		},
		Input{},
		true,
	},
}
