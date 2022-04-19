package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCurrentStatsSms(t *testing.T) {
	stats := ValidatorStats{}
	vm := ValidatorMonitor{}

	result := getCurrentStatsSms(stats, &vm)
	assert.Equal(t, " (N/A% up) \n 🔴 Height **N/A**\n🟡 Latest Blocks Signed: **N/A**", result)
	println(result)
}
