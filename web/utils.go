package web

import (
	"math"

	"github.com/dustin/go-humanize"
	"github.com/montanaflynn/stats"
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
	val, _ := stats.Round(n/math.Pow(1000, float64(exp)), 2)
	return humanize.Commaf(val) + " " + suffix
}

func humanizeCommaf(f float64) string {
	val, _ := stats.Round(f, 0)
	return humanize.Commaf(val)
}

func humanizeVolume(f float64) string {
	if float64(int64(f)) != f {
		return humanize.Commaf(f)
	}
	val, _ := stats.Round(f, 0)
	return humanize.Commaf(val)
}
