package staticdump

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

// NOTE(sudorandom): This entire file seems incresibly bad. For some reason, CCP stopped updating the static data to coorespond to the
// item renames that they started doing this year.

var nameOverrides = map[int64]string{
	// https://www.eveonline.com/article/q46atf/patch-notes-for-january-2020-release
	4533: "Small ACM Compact Armor Repairer",
	4529: "Small I-a Enduring Armor Repairer",
	4573: "Medium ACM Compact Armor Repairer",
	4569: "Medium I-a Enduring Armor Repairer",
	4613: "Large ACM Compact Armor Repairer",
	4609: "Large I-a Enduring Armor Repairer",
	4531: "Small ACM Compact Armor Repairer",
	4535: "Small I-a Enduring Armor Repairer",
	4571: "Medium ACM Compact Armor Repairer",
	4575: "Medium I-a Enduring Armor Repairer",
	4579: "'Meditation' Medium Armor Repairer I",
	4611: "Large ACM Compact Armor Repairer",
	4615: "Large I-a Enduring Armor Repairer",

	// https://www.eveonline.com/article/q5jak4/patch-notes-for-february-2020-release
	5093:  "Small Radiative Scoped Remote Capacitor Transmitter",
	5091:  "Small Inductive Compact Remote Capacitor Transmitter",
	16489: "Medium Radiative Scoped Remote Capacitor Transmitter",
	16495: "Medium Inductive Compact Remote Capacitor Transmitter",
	16481: "Large Radiative Scoped Remote Capacitor Transmitter",
	16487: "Large Inductive Compact Remote Capacitor Transmitter",

	5087:  "Small Radiative Scoped Remote Capacitor Transmitter",
	5089:  "Small Inductive Compact Remote Capacitor Transmitter",
	16493: "Medium Radiative Scoped Remote Capacitor Transmitter",
	16491: "Medium Inductive Compact Remote Capacitor Transmitter",
	16485: "Large Radiative Scoped Remote Capacitor Transmitter",
	16483: "Large Inductive Compact Remote Capacitor Transmitter",

	// https://www.eveonline.com/article/q6z2qy/patch-notes-for-march-2020-release
	4959: "'Seed' Micro Capacitor Booster I",
	5011: "Small F-RX Compact Capacitor Booster",
	4833: "Medium F-RX Compact Capacitor Booster",
	5051: "Heavy F-RX Compact Capacitor Booster",

	4957:  "'Seed' Micro Capacitor Booster I",
	4961:  "'Seed' Micro Capacitor Booster I",
	4955:  "'Seed' Micro Capacitor Booster I",
	3556:  "'Seed' Micro Capacitor Booster I",
	3558:  "'Seed' Micro Capacitor Booster I",
	15774: "'Seed' Micro Capacitor Booster I",
	14180: "'Seed' Micro Capacitor Booster I",
	14182: "'Seed' Micro Capacitor Booster I",
	15782: "'Seed' Micro Capacitor Booster I",
	5009:  "Small F-RX Compact Capacitor Booster",
	5013:  "Small F-RX Compact Capacitor Booster",
	5007:  "Small F-RX Compact Capacitor Booster",
	4831:  "Medium F-RX Compact Capacitor Booster",
	4835:  "Medium F-RX Compact Capacitor Booster",
	4829:  "Medium F-RX Compact Capacitor Booster",
	5049:  "Heavy F-RX Compact Capacitor Booster",
	5053:  "Heavy F-RX Compact Capacitor Booster",
	5047:  "Heavy F-RX Compact Capacitor Booster",

	// https://www.eveonline.com/article/q8tteh/patch-notes-for-18-04-release
	578:  "Adaptive Invulnerability Shield Hardener",
	2293: "Anti-EM Shield Hardener",
	2289: "Anti-Explosive Shield Hardener",
	2291: "Anti-Kinetic Shield Hardener",
	2295: "Anti-Thermal Shield Hardener",
	9632: "Compact Adaptive Invulnerability Shield Hardener",
	9622: "Compact Anti-EM Shield Hardener",
	9646: "Compact Anti-Explosive Shield Hardener",
	9608: "Compact Anti-Kinetic Shield Hardener",
	9660: "Compact Anti-Thermal Shield Hardener",

	// https://www.eveonline.com/article/qaxqcl/patch-notes-for-version-18-05
	// patch_18_04ArmorHardenerOverrideRegex handles "FLAVORNAME Armor DAMAGETYPE Hardener → FLAVORNAME DAMAGETYPE Armor Hardener."
	16357: "Experimental Enduring EM Armor Hardener I",
	16365: "Experimental Enduring Explosive Armor Hardener I",
	16373: "Experimental Enduring Kinetic Armor Hardener I",
	16381: "Experimental Enduring Thermal Armor Hardener I",
	// Prototype Compact
	16359: "Prototype Compact EM Armor Hardener I",
	16367: "Prototype Compact Explosive Armor Hardener I",
	16375: "Prototype Compact Kinetic Armor Hardener I",
	16383: "Prototype Compact Thermal Armor Hardener I",
}

func computeAliases(typeID int64, typeName string) (string, []string) {
	nameOverride, ok := nameOverrides[typeID]
	if ok {
		// log.Printf("%d %s -> %s", typeID, typeName, nameOverride)
		return nameOverride, []string{typeName}
	}

	armorHwOverride, ok := patch_18_04ArmorHardenerOverride(typeName)
	if ok {
		// log.Printf("%d %s -> %s", typeID, typeName, armorHwOverride)
		return armorHwOverride, []string{typeName}
	}

	rigOverride, ok := patch_18_04RigOverride(typeName)
	if ok {
		// log.Printf("%d %s -> %s", typeID, typeName, rigOverride)
		return rigOverride, []string{typeName}
	}

	return typeName, nil
}

var sizes = []string{
	"Small",
	"Medium",
	"Large",
	"Capital",
}

var damageTypes = []string{
	"EM",
	"Kinetic",
	"Explosive",
	"Thermal",
}

var patch_18_04RigOverrideRegex = regexp.MustCompile(fmt.Sprintf(
	`^(%s) Anti-(%s) (Screen Reinforcer|Pump) (I|II)$`,
	strings.Join(sizes, "|"),
	strings.Join(damageTypes, "|"),
))

func patch_18_04RigOverride(typeName string) (string, bool) {
	match := patch_18_04RigOverrideRegex.FindStringSubmatch(typeName)
	if match == nil {
		return "", false
	}

	b := bytes.NewBuffer(nil)
	// Size
	b.WriteString(match[1])
	b.WriteString(" ")
	// Damage Type
	b.WriteString(match[2])
	b.WriteString(" ")
	switch match[3] {
	case "Screen Reinforcer":
		b.WriteString("Shield Reinforcer")
	case "Pump":
		b.WriteString("Armor Reinforcer")
	}
	b.WriteString(" ")
	b.WriteString(match[4])
	return b.String(), true
}

var flavorTypes = []string{
	"",
	"Experimental",
	"Limited",
	"Prototype",
	"Upgraded",
	"Ammatar Navy",
	"Dark Blood",
	"Domination",
	"Federation Navy",
	"Imperial Navy",
	"Khanid Navy",
	"Republic Fleet",
	"Shadow Serpentis",
	"True Sansha",
	"Ahremen's Modified",
	"Brokara's Modified",
	"Brynn's Modified",
	"Chelm's Modified",
	"Cormack's Modified",
	"Draclira's Modified",
	"Raysere's Modified",
	"Selynne's Modified",
	"Setele's Modified",
	"Tairei's Modified",
	"Tuvan's Modified",
	"Vizan's Modified",
	"Centus A-Type",
	"Centus B-Type",
	"Centus C-Type",
	"Centus X-Type",
	"Core A-Type",
	"Core B-Type",
	"Core C-Type",
	"Core X-Type",
	"Corpus A-Type",
	"Corpus B-Type",
	"Corpus C-Type",
	"Corpus X-Type",
}

var patch_18_04ArmorHardenerOverrideRegex = regexp.MustCompile(fmt.Sprintf(
	`^(%s) Armor (%s) Hardener ?(I|II)? ?(Blueprint)?$`,
	strings.Join(flavorTypes, "|"),
	strings.Join(damageTypes, "|"),
))

func patch_18_04ArmorHardenerOverride(typeName string) (string, bool) {
	match := patch_18_04ArmorHardenerOverrideRegex.FindStringSubmatch(typeName)
	if match == nil {
		return "", false
	}

	b := bytes.NewBuffer(nil)
	// Flavor
	if match[1] != "" {
		b.WriteString(match[1])
		b.WriteString(" ")
	}
	// Damage Type
	b.WriteString(match[2])

	b.WriteString(" Armor Hardener")

	// Level
	if match[3] != "" {
		b.WriteString(" ")
		b.WriteString(match[3])
	}
	// Blueprint
	if match[4] != "" {
		b.WriteString(" ")
		b.WriteString(match[4])
	}

	return b.String(), true
}
