package parsers

var miningLedgerTestCases = []Case{
	{
		"Example",
		`Timestamp	Ore Type	Quantity	Volume	Est. Price	Solar System	Ore TypeID	SolarSystemID
2017.10.11	Triclinic Bistot	349	5584.0	962,161	Z-Y9C3	17428	30003124
2017.10.22	Crimson Arkonor	435	6960	2,253,447	Z-Y9C3	17425	30003124
2017.10.05	Bright Spodumain	865	13840.0	1,999,603	Z-Y9C3	17466	30003124
2017.10.03	Triclinic Bistot	1606	25696.0	4,427,597	Z-Y9C3	17428	30003124
2017.10.03	Bright Spodumain	1638	26208.0	3,786,531	Z-Y9C3	17466	30003124
2017.09.27	Triclinic Bistot	1919	30704.0	5,290,510	Z-Y9C3	17428	30003124
2017.10.04	Bright Spodumain	2793	44688.0	6,456,522	Z-Y9C3	17466	30003124
2017.10.22	Bright Spodumain	2986	47776.0	6,902,676	Z-Y9C3	17466	30003124
2017.10.18	Crimson Arkonor	3167	50672	16,406,136	Z-Y9C3	17425	30003124
2017.10.13	Bright Spodumain	3822	61152.0	8,835,240	Z-Y9C3	17466	30003124
2017.09.29	Bright Spodumain	4371	69936.0	10,104,353	Z-Y9C3	17466	30003124
2017.10.11	Monoclinic Bistot	6628	106048.0	18,576,627	Z-Y9C3	17429	30003124
2017.10.05	Triclinic Bistot	6702	107232.0	18,476,810	Z-Y9C3	17428	30003124
2017.09.27	Bright Spodumain	6775	108400.0	15,661,631	Z-Y9C3	17466	30003124
2017.09.27	Iridescent Gneiss	8752	43760.0	10,599,547	Z-Y9C3	17865	30003124
2017.10.04	Iridescent Gneiss	9072	45360.0	10,987,099	Z-Y9C3	17865	30003124
2017.10.05	Iridescent Gneiss	9279	46395.0	11,237,796	Z-Y9C3	17865	30003124
2017.10.18	Bright Spodumain	9962	159392.0	23,028,956	Z-Y9C3	17466	30003124
2017.10.13	Iridescent Gneiss	13074	65370.0	15,833,921	Z-Y9C3	17865	30003124`,
		&MiningLedger{
			Items: []MiningLedgerItem{
				{Name: "Bright Spodumain", Quantity: 33212},
				{Name: "Crimson Arkonor", Quantity: 3602},
				{Name: "Iridescent Gneiss", Quantity: 40177},
				{Name: "Monoclinic Bistot", Quantity: 6628},
				{Name: "Triclinic Bistot", Quantity: 10576},
			},
			lines: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19},
		},
		Input{},
		true,
	},
}
