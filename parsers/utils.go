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

func regexParseLines(re *regexp.Regexp, lines []string) ([][]string, []int, []string) {
	var matches [][]string
	var matchedLines []int
	var rest []string
	for i, line := range lines {
		match := re.FindStringSubmatch(line)
		if len(match) == 0 {
			rest = append(rest, line)
		} else {
			matches = append(matches, match)
			matchedLines = append(matchedLines, i)
		}
	}
	return matches, matchedLines, rest
}
