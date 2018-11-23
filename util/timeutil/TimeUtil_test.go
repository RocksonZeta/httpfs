package timeutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	assert.Equal(t, "00:10", FormatDuration(10))
}
