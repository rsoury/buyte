package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRjust(t *testing.T) {
	assert := assert.New(t)

	assert.Equal(Rjust("7", 2, "0"), "07", "Original value should be prepended, if the length is less than the limit provided.")
}
