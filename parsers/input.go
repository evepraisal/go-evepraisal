package parsers

import (
	"bytes"
	"strings"
)

// Input is used as the input to parsers. This exists so that the first part of parsing the text isn't duplicated for each parser
type Input map[int]string

// StringsToInput converts an array of strings into an Input object
func StringsToInput(lines []string) Input {
	m := make(Input)
	for i, line := range lines {
		m[i] = line
	}
	return m
}

// StringToInput converts a strings into an Input object
func StringToInput(s string) Input {
	s = strings.Replace(s, "\r", "", -1)
	return StringsToInput(strings.Split(s, "\n"))
}

// Strings returns an array of strings from an Input object
func (m Input) Strings() []string {
	keys := make([]int, 0)
	for k := range m {
		keys = append(keys, k)
	}

	lines := make([]string, len(keys))
	i := 0
	for k := range keys {
		lines[i] = m[k]
		i++
	}
	return lines
}

// String returns a string from an Input object
func (m Input) String() string {
	var buffer bytes.Buffer
	for _, line := range m.Strings() {
		buffer.WriteString(line)
		buffer.WriteByte('\n')
	}
	return buffer.String()
}
