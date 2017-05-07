package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

type AssetList struct {
	items []AssetItem
	lines []int
}

func (r *AssetList) Name() string {
	return "assets"
}

func (r *AssetList) Lines() []int {
	return r.lines
}

type AssetItem struct {
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

var reAssetList = regexp.MustCompile(strings.Join([]string{
	`^([\S\ ]*)`,                           // name
	`\t([\d,'\.]*)`,                        // quantity
	`(?:\t([\S ]*))?`,                      // group
	`(?:\t([\S ]*))?`,                      // category
	`(?:\t(XLarge|Large|Medium|Small|))?`,  // size
	`(?:\t(High|Medium|Low|Rigs|[\d ]*))?`, // slot
	`(?:\t([\d ,\.]*) m3)?`,                // volume
	`(?:\t([\d]+|))?`,                      // meta level
	`(?:\t([\d]+|))?$`,                     // tech level
}, ""))

func ParseAssets(input Input) (ParserResult, Input) {
	assetList := &AssetList{}
	matches, rest := regexParseLines(reAssetList, input)
	assetList.lines = regexMatchedLines(matches)
	for _, match := range matches {
		assetList.items = append(assetList.items,
			AssetItem{
				name:      match[1],
				quantity:  ToInt(match[2]),
				volume:    ToFloat64(match[7]),
				group:     match[3],
				category:  match[4],
				size:      match[5],
				slot:      match[6],
				metaLevel: match[8],
				techLevel: match[9],
			})
	}
	sort.Slice(assetList.items, func(i, j int) bool {
		return fmt.Sprintf("%v", assetList.items[i]) < fmt.Sprintf("%v", assetList.items[j])
	})
	return assetList, rest
}
