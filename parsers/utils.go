package parsers

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
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

func CleanString(s string) string {
	return strings.Trim(s, " ")
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
