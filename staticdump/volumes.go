package staticdump

// Reference: https://bitbucket.org/snippets/viktorielucilla/d4oyA/calculating-repackaged-volumes-for-eve
var volumeGroupOverrides = map[int64]float64{
	25:   2500,     // Frigate
	26:   10000,    // Cruiser
	27:   50000,    // Battleship
	28:   20000,    // Industrial
	30:   10000000, // Titan
	31:   500,      // Shuttle
	237:  2500,     // Rookie ship
	324:  2500,     // Assault Frigate
	358:  10000,    // Heavy Assault Cruiser
	380:  20000,    // Deep Space Transport
	419:  15000,    // Combat Battlecruiser
	420:  5000,     // Destroyer
	463:  3750,     // Mining Barge
	485:  1300000,  // Dreadnought
	513:  1300000,  // Freighter
	540:  15000,    // Command Ship
	541:  5000,     // Interdictor
	543:  3750,     // Exhumer
	547:  1300000,  // Carrier
	659:  1300000,  // Supercarrier
	830:  2500,     // Covert Ops
	831:  2500,     // Interceptor
	832:  10000,    // Logistics
	833:  10000,    // Force Recon Ship
	834:  2500,     // Stealth Bomber
	883:  1300000,  // Capital Industrial Ship
	893:  2500,     // Electronic Attack Ship
	894:  10000,    // Heavy Interdiction Cruiser
	898:  50000,    // Black Ops
	900:  50000,    // Marauder
	902:  1300000,  // Jump Freighter
	906:  10000,    // Combat Recon Ship
	941:  500000,   // Industrial Command Ship
	963:  5000,     // Strategic Cruiser
	1022: 500,      // Prototype Exploration Ship
	1201: 15000,    // Attack Battlecruiser
	1202: 20000,    // Blockade Runner
	1283: 2500,     // Expedition Frigate
	1305: 5000,     // Tactical Destroyer
	1527: 2500,     // Logistics Frigate
	1534: 5000,     // Command Destroyer
	1538: 1300000,  // Force Auxiliary

	// Modules
	1815: 8000, // Phenomena Generators
}

var volumeMarketGroupOverrides = map[int64]float64{
	1614: 20000, // Special Edition Industrial Ships
	1056: 1000,  // Capital Ship Repaierers
	801:  1000,  // Siege Modules
	600:  1000,
	771:  1000,
	772:  1000,
	773:  1000,
	774:  1000,
	775:  1000,
	776:  1000,
	777:  1000,
	778:  1000,
	910:  1000,
	1052: 1000,
	1063: 1000,
	2240: 1000,
	2241: 1000,
	2242: 1000,
	2243: 1000,
	2244: 1000,
	2245: 1000,
	2246: 1000,
	2247: 1000,
	2250: 1000,
	2251: 1000,
	2249: 2000,
	2267: 2000,
	2268: 2000,
	2269: 2000,
	2276: 2000,
}

var volumeItemOverrides = map[int64]float64{
	42244: 50000,  // Porpoise
	17363: 10,     // Small Container
	3467:  10,     // Small Container
	33011: 10,     // Small Container
	3297:  10,     // Small Container
	17364: 33,     // Medium Container
	3466:  33,     // Medium Container
	3293:  33,     // Medium Container
	33009: 50,     // Medium Freight Container
	17365: 65,     // Large Container
	3465:  65,     // Large Container
	3296:  65,     // Large Container
	33007: 100,    // Large Freight Container
	11488: 150,    // Huge Secure Container
	33005: 500,    // Huge Freight Container
	11489: 300,    // Giant Secure Container
	24445: 1200,   // Giant Freight Container
	33003: 2500,   // Enormous Freight Container
	17366: 10000,  // Station Container
	17367: 50000,  // Station Vault Container
	17368: 100000, // Station Warehouse Container
	41237: 1000,   // 10000MN Y-S8 Compact Afterburner
	41417: 1000,   // Sentient Fighter Support Unit
	41249: 1000,
	41250: 1000,
	41251: 1000,
	41252: 1000,
	41253: 1000,
	41254: 1000,
	41255: 1000,
	41236: 1000,
	41238: 1000,
	41239: 1000,
	41240: 1000,
	41241: 1000,
	41411: 1000,
	24283: 1000,
	41414: 1000,
	41415: 1000,
	40715: 2000,
	40716: 2000,
	40717: 2000,
	40718: 2000,
	40714: 2000,
	45534: 10000,
}
