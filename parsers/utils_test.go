package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInt(t *testing.T) {
	assert.Equal(t, int64(0), ToInt(""))
	assert.Equal(t, int64(42), ToInt("42"))
}

func TestToFloat64(t *testing.T) {
	assert.Equal(t, float64(37805997.92), ToFloat64("37,805,997.92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37 805 997,92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37 805 997.92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37 805 997,92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37'805'997,92"))
	assert.Equal(t, float64(0.02), ToFloat64("0,02"))
}
