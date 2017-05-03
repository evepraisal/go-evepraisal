package parsers

import (
	"regexp"
	"strings"
)

type AssetRow struct {
	name      string
	quantity  int64
	volume    float64
	group     string
	category  string
	size      string
	slot      string
	metaLevel string
	techLevel string
}

func (r AssetRow) Name() string {
	return r.name
}

func (r AssetRow) Quantity() int64 {
	return r.quantity
}

func (r AssetRow) Volume() float64 {
	return r.volume
}

var reAssetList = regexp.MustCompile(strings.Join([]string{
	`^([\S\ ]*)`,                         // name
	`\t([\d,'\.]*)`,                      // quantity
	`(\t([\S ]*))?`,                      // group
	`(\t([\S ]*))?`,                      // category
	`(\t(XLarge|Large|Medium|Small|))?`,  // size
	`(\t(High|Medium|Low|Rigs|[\d ]*))?`, // slot
	`(\t([\d ,\.]*) m3)?`,                // volume
	`(\t([\d]+|))?`,                      // meta level
	`(\t([\d]+|))?$`,                     // tech level
}, ""))

func ParseAssets(lines []string) ([]ParserResult, []string) {
	var results []ParserResult
	matches, rest := regexParseLines(reAssetList, lines)
	for _, match := range matches {
		results = append(results,
			AssetRow{
				name:      match[1],
				quantity:  ToInt(match[2]),
				volume:    ToFloat64(match[12]),
				group:     match[4],
				category:  match[6],
				size:      match[8],
				slot:      match[10],
				metaLevel: match[14],
				techLevel: match[16],
			})
	}
	return results, rest
}
