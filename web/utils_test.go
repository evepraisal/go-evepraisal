package web

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHumanLargeNumber(t *testing.T) {
	assert.Equal(t, HumanLargeNumber(100.0), "100")
	assert.Equal(t, HumanLargeNumber(123000000), "123 Million")
	assert.Equal(t, HumanLargeNumber(99999999), "100 Million")
	assert.Equal(t, HumanLargeNumber(999999999999999999999999999999), "1 Nonillion")
}
