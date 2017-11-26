package web

import (
	"math"

	"github.com/dustin/go-humanize"
	"github.com/evepraisal/go-evepraisal"
	"github.com/leekchan/accounting"
	"github.com/montanaflynn/stats"
)

// ISKFormat defines how ISK is rounded and displayed
var ISKFormat = accounting.Accounting{Symbol: "", Precision: 2}

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

// HumanLargeNumber returns string numbers that are formatted with commas and have X-illion appended.
func HumanLargeNumber(n float64) string {
	if math.Abs(n) < 1000 {
		return humanize.Commaf(n)
	}

	exp := int((math.Log(math.Abs(n)) / math.Log(1000)))
	suffix := humanThresholds[int(math.Min(float64(exp-1), 10))]
	val, err := stats.Round(n/math.Pow(1000, float64(exp)), 2)
	if err != nil {
		return humanize.Commaf(n)
	}
	return humanize.Commaf(val) + " " + suffix
}

func humanizeCommaf(f float64) string {
	return ISKFormat.FormatMoney(f)
}

func humanizeVolume(f float64) string {
	if float64(int64(f)) != f {
		return humanize.Commaf(f)
	}
	val, err := stats.Round(f, 0)
	if err != nil {
		return humanize.Commaf(f)
	}
	return humanize.Commaf(val)
}

func cleanAppraisal(appraisal *evepraisal.Appraisal) *evepraisal.Appraisal {
	appraisal.User = nil
	return appraisal
}

func cleanAppraisals(appraisals []evepraisal.Appraisal) []evepraisal.Appraisal {
	cleanedAppraisals := make([]evepraisal.Appraisal, len(appraisals))
	for i, a := range appraisals {
		cleanedAppraisal := cleanAppraisal(&a)
		cleanedAppraisals[i] = *cleanedAppraisal
	}
	return cleanedAppraisals
}
