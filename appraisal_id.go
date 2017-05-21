package evepraisal

import "github.com/martinlindhe/base36"

func AppraisalIDToUint64(appraisalID string) uint64 {
	return base36.Decode(appraisalID)
}

func Uint64ToAppraisalID(aID uint64) string {
	return base36.Encode(aID)
}
