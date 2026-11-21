package identity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAUTH_04_PolicyWhitespace(t *testing.T) {
	assert.ErrorIs(t, ValidatePassword("            "), ErrPasswordWhitespace)
}
