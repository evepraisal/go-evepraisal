package parsers

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var bigNumberRegex = `[\d,'\.\ ’` + "\u00a0\xc2\xa0’" + `]`
var cleanIntegers = regexp.MustCompile(`[,\'\.\ ’` + "\u00a0\xc2\xa0" + `]`)
var separatorCharacters = map[rune]bool{
	',':    true,
	'.':    true,
	' ':    true,
	'\'':   true,
	'\xc2': true,
	'\xa0': true,
	'’':    true,
}

func splitDecimal(s string) (string, string) {
	runes := []rune(s)
	if len(runes) > 3 {
		_, twodecimal := separatorCharacters[runes[len(runes)-3]]
		if twodecimal {
			whole := string(runes[0 : len(runes)-3])
			decimal := string(runes[len(runes)-2:])
			return whole, decimal
		}
	}
	if len(runes) > 2 {
		_, onedecimal := separatorCharacters[runes[len(runes)-2]]
		if onedecimal {
			whole := string(runes[0 : len(runes)-2])
			decimal := string(runes[len(runes)-1:])
			return whole, decimal
		}
	}

	return s, ""
}

// ToInt parses a string into an integer. It will return 0 on failure
func ToInt(s string) int64 {
	if s == "" {
		return 0
	}

	whole, _ := splitDecimal(s)
	cleaned := cleanIntegers.ReplaceAllString(whole, "")

	i, err := strconv.ParseInt(cleaned, 10, 64)
	if err == nil {
		return i
	}

	return 0
}

// ToFloat64 parses a string into a float64. It will return 0.0 on failure
func ToFloat64(s string) float64 {
	// Attempt to parse float as "normal"
	f, err := strconv.ParseFloat(s, 64)
	if err == nil {
		return f
	}

	whole, decimal := splitDecimal(s)
	f, _ = strconv.ParseFloat(fmt.Sprintf("%d.%s", ToInt(string(whole)), string(decimal)), 64)

	return f
}

// CleanTypeName will remove leading and trailing whitespace and leading asterisks.
func CleanTypeName(s string) string {
	return strings.TrimSuffix(strings.Trim(s, " "), "*")
}

func regexMatchedLines(matches map[int][]string) []int {
	keys := make([]int, len(matches))
	i := 0
	for k := range matches {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	return keys
}

func regexParseLines(re *regexp.Regexp, input Input) (map[int][]string, Input) {
	matches := make(map[int][]string)
	rest := make(Input)
	for i, line := range input {
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			rest[i] = line
		} else {
			matches[i] = match
		}
	}
	return matches, rest
}
