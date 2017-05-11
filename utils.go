package evepraisal

import (
	"math"
	"strings"

	"github.com/dustin/go-humanize"
)

var humanThresholds = []string{
	"Thousand",
	"Million",
	"Billion",
	"Trillion",
	"Quadrillion",
	"Quintillion",
	"Sextillion",
	"Septillion",
	"Octillion",
	"Nonillion",
	"Decillion",
}

func HumanLargeNumber(n float64) string {
	if math.Abs(n) < 1000 {
		return humanize.Commaf(n)
	}

	exp := int((math.Log(math.Abs(n)) / math.Log(1000)))
	suffix := humanThresholds[int(math.Min(float64(exp-1), 10))]
	numberStr := TruncateFloatDigits(humanize.Commaf(n / math.Pow(1000, float64(exp))))
	return numberStr + " " + suffix
}

func TruncateFloatDigits(s string) string {
	parts := strings.Split(s, ".")
	if len(parts) == 1 {
		return s
	}
	belowZero := parts[1][0:2]
	if belowZero == "00" {
		return parts[0]
	}

	return parts[0] + "." + belowZero
}
