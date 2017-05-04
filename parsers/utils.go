package parsers

import (
	"regexp"
	"strconv"
)

var cleanIntegers = regexp.MustCompile(`[,\'\.` + "\xc2\xa0" + `]`)

func ToInt(s string) int64 {
	if s == "" {
		return 0
	}

	s = cleanIntegers.ReplaceAllString(s, "")

	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return int64(ToFloat64(s))
	}
	return i
}

func ToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return f
}

func regexParseLines(re *regexp.Regexp, lines []string) ([][]string, []string, []string) {
	var matches [][]string
	var raw []string
	var rest []string
	for _, line := range lines {
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			rest = append(rest, line)
		} else {
			matches = append(matches, match)
			raw = append(raw, line)
		}
	}
	return matches, raw, rest
}
