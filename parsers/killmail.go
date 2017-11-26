package parsers

import (
	"fmt"
	"regexp"
	"strings"
)

// Killmail is the result from the killmail parser
type Killmail struct {
	Datetime  string
	Victim    map[string]interface{}
	Involved  []map[string]interface{}
	Destroyed []KillmailItem
	Dropped   []KillmailItem
	lineCount int
}

// Name returns the parser name
func (r *Killmail) Name() string {
	return "killmail"
}

// Lines returns the lines that this result is made from
func (r *Killmail) Lines() []int {
	lines := make([]int, r.lineCount)
	for i := 0; i < r.lineCount; i++ {
		lines[i] = i
	}
	return lines
}

// KillmailItem is a single item from a killmail result
type KillmailItem struct {
	Name     string
	Quantity int64
	Location string
}

var reKillmailDate = regexp.MustCompile(`^(\d\d\d\d.\d\d.\d\d \d\d:\d\d(:\d\d)?)$`)
var reKillmailPlayerLine = regexp.MustCompile(`^([\w\s]+): ([\S ]+)$`)
var reKillmailInvolvedLine = regexp.MustCompile(`^([\w ]+): ([\S ]+?)( \(laid the final blow\))?$`)
var reKillmailItemLine = regexp.MustCompile(`^([\w '-]+?)(?:, Qty: (\d+))?(?: \(([\w ]+)\))?$`)

// ParseKillmail parses a killmail
func ParseKillmail(input Input) (ParserResult, Input) {
	killmail := &Killmail{}
	if len(input) == 0 {
		return nil, input
	}

	killmail.lineCount = len(input)

	// First line should be the datestamp
	dateParts := reKillmailDate.FindStringSubmatch(input[0])
	if len(dateParts) == 0 {
		return nil, input
	}
	killmail.Datetime = dateParts[1]

	var err error
	inputLines := input.Strings()
	offset := 2
	victim, victimOffset, err := parsekillmailVictim(inputLines[offset:])
	if err != nil {
		return nil, input
	}
	killmail.Victim = victim
	offset += victimOffset

	// skip past the blank line
	offset++

	for ; offset < len(inputLines); offset++ {
		line := inputLines[offset]
		if line == "Involved parties:" {
			offset += 2
			involved, involvedOffset, err := parsekillmailInvolved(inputLines[offset:])
			if err != nil {
				return nil, input
			}
			killmail.Involved = involved
			offset += involvedOffset
		} else if line == "Destroyed items:" {
			offset += 2
			destroyed, destroyedOffset, err := parsekillmailItems(inputLines[offset:])
			if err != nil {
				return nil, input
			}
			killmail.Destroyed = destroyed
			offset += destroyedOffset
		} else if line == "Dropped items:" {
			offset += 2
			dropped, droppedOffset, err := parsekillmailItems(inputLines[offset:])
			if err != nil {
				return nil, input
			}
			killmail.Dropped = dropped
			offset += droppedOffset
		} else {
			return nil, input
		}
	}

	return killmail, nil
}

func parsekillmailVictim(lines []string) (map[string]interface{}, int, error) {
	victim := make(map[string]interface{})
	var i int
	var line string
	for i, line = range lines {
		if line == "" {
			return victim, i, nil
		}
		result := reKillmailPlayerLine.FindStringSubmatch(line)
		if len(result) == 0 {
			return victim, i, fmt.Errorf("Cannot parse victim data (line %d)", i)
		}
		victim[killmailKeyFormat(result[1])] = result[2]
	}
	return victim, i, nil
}

func parsekillmailInvolved(lines []string) ([]map[string]interface{}, int, error) {
	involved := make([]map[string]interface{}, 0)
	var i int
	var line string
	player := make(map[string]interface{})
	for i, line = range lines {
		if line == "" {
			if len(lines) > i+1 && strings.HasPrefix(lines[i+1], "Name:") {
				involved = append(involved, player)
				player = make(map[string]interface{})
				continue
			}

			break
		}

		result := reKillmailInvolvedLine.FindStringSubmatch(line)
		if len(result) == 0 {
			return involved, i, fmt.Errorf("Cannot parse involved data (line %d)", i)
		}

		if len(result[3]) > 0 {
			player["killing_blow"] = true
		}

		player[killmailKeyFormat(result[1])] = result[2]
	}

	if len(player) != 0 {
		involved = append(involved, player)
	}

	return involved, i, nil
}

func parsekillmailItems(lines []string) ([]KillmailItem, int, error) {
	items := make([]KillmailItem, 0)
	var i int
	var line string
	for i, line = range lines {
		if line == "" {
			return items, i, nil
		}

		result := reKillmailItemLine.FindStringSubmatch(line)
		if len(result) == 0 {
			return items, i, fmt.Errorf("Cannot parse items data (line %d)", i)
		}

		quantity := int64(1)
		if result[2] != "" {
			quantity = ToInt(result[2])
		}
		items = append(items, KillmailItem{
			Name:     result[1],
			Quantity: quantity,
			Location: result[3],
		})
	}
	return items, i, nil
}

func killmailKeyFormat(s string) string {
	return strings.Replace(strings.ToLower(s), " ", "_", -1)
}
