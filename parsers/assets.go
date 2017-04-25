package parsers

import (
	"log"
	"regexp"
	"strings"
)

type AssetRow struct {
	Name      string
	Quantity  int64
	Group     string
	Category  string
	Size      string
	Slot      string
	Volume    float64
	MetaLevel string
	TechLevel string
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

func ParseAssets(lines []string) ([]IResult, []string) {
	var matches []IResult
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
					match[4],
					match[6],
					match[8],
					match[10],
					ToFloat64(match[12]),
					match[14],
					match[16],
				})
		}
	}
	return matches, rest
}
