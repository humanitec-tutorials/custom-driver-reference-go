package api

import (
	"testing"

	"gotest.tools/assert"
)

func TestValidIDs(t *testing.T) {
	assert.Assert(t, isValidAsID("valid-id"))
	assert.Assert(t, isValidAsID("01-valid-id-2"))
	assert.Assert(t, isValidAsID("jahgsdo87iq28ui3hdgkuyqxl3"))

	assert.Assert(t, !isValidAsID("-invalid-id"))
	assert.Assert(t, !isValidAsID("Invalid ID"))
	assert.Assert(t, !isValidAsID(""))
	assert.Assert(t, !isValidAsID("a"))
}
