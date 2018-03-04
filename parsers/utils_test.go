package parsers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToInt(t *testing.T) {
	assert.Equal(t, int64(0), ToInt(""))
	assert.Equal(t, int64(42), ToInt("42"))
	assert.Equal(t, int64(3225), ToInt("3,225"))
	assert.Equal(t, int64(10), ToInt("10.1"))
	assert.Equal(t, int64(10), ToInt("10.11"))
	assert.Equal(t, int64(1667), ToInt("1\u00a0667"))

	// NOTE: This might be considered a flaw, but this is a limitation because we don't know the locale
	assert.Equal(t, int64(10111), ToInt("10.111"))
}

func TestToFloat64(t *testing.T) {
	assert.Equal(t, float64(37805997.92), ToFloat64("37,805,997.92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37 805 997,92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37 805 997.92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37 805 997,92"))
	assert.Equal(t, float64(37805997.92), ToFloat64("37'805'997,92"))
	assert.Equal(t, float64(0.02), ToFloat64("0,02"))
	assert.Equal(t, float64(3225), ToFloat64("3\u00a0225"))
	assert.Equal(t, float64(3225.2), ToFloat64("3'225.20"))
	assert.Equal(t, float64(1234.2), ToFloat64("1,234.2"))
	assert.Equal(t, float64(1667), ToFloat64("1\u00a0667"))
}
