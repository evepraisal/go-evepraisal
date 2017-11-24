package evepraisal

import "github.com/martinlindhe/base36"

// AppraisalIDToUint64 returns a Uint64 appraisal ID for the given string version appraisalID.
// This is used to make nicer/smaller appraisal IDs
func AppraisalIDToUint64(appraisalID string) uint64 {
	return base36.Decode(appraisalID)
}

// Uint64ToAppraisalID returns a string AppraisalID for the given Uint64 representation
func Uint64ToAppraisalID(aID uint64) string {
	return base36.Encode(aID)
}
