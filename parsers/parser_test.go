package parsers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Case struct {
	Description  string
	Input        string
	Expected     []ParserResult
	ExpectedRest []string
}

type CaseGroup struct {
	name  string
	funct func(lines []string) ([]ParserResult, []string)
	cases []Case
}

var ParserTests = []CaseGroup{
	CaseGroup{"assets", ParseAssets, assetListTestCases},
}

func TestParsers(rt *testing.T) {
	for _, group := range ParserTests {
		for _, c := range group.cases {
			rt.Run(group.name+c.Description, func(t *testing.T) {
				result, rest := group.funct(strings.Split(c.Input, "\n"))
				assert.Equal(t, c.Expected, result, "results should be the same")
				assert.Equal(t, c.ExpectedRest, rest, "the rest should be the same")
			})
		}
	}
}
