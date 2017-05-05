package parsers

import (
	"bytes"
	"strings"
)

type Input map[int]string

func StringToInput(input string) Input {
	m := make(Input)
	for i, line := range strings.Split(input, "\n") {
		m[i] = line
	}
	return m
}

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

func (m Input) String() string {
	var buffer bytes.Buffer
	for _, line := range m.Strings() {
		buffer.WriteString(line)
		buffer.WriteByte('\n')
	}
	return buffer.String()
}
