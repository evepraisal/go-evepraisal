package parsers

import (
	"log"
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

var assetList = regexp.MustCompile(strings.Join([]string{
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
	var matches []ParserResult
	var rest []string
	for _, line := range lines {
		match := assetList.FindStringSubmatch(line)
		log.Printf("%#v", match)
		if len(match) == 0 {
			rest = append(rest, line)
		} else {
			matches = append(matches,
				AssetRow{
					match[1],
					ToInt(match[2]),
					ToFloat64(match[12]),
					match[4],
					match[6],
					match[8],
					match[10],
					match[14],
					match[16],
				})
		}
	}
	return matches, rest
}
