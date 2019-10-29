package perm_test

import (
	"testing"

	"github.com/enpointe/activity/perm"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	assert.Equal(t, perm.Admin.String(), "admin")
	assert.Equal(t, perm.Staff.String(), "staff")
	assert.Equal(t, perm.Basic.String(), "basic")
}

func TestConvert(t *testing.T) {
	assert.Equal(t, perm.Convert("admin"), perm.Admin)
	assert.Equal(t, perm.Convert("staff"), perm.Staff)
	assert.Equal(t, perm.Convert("basic"), perm.Basic)
}

func TestAuthorized(t *testing.T) {
	assert.True(t, perm.Admin.Grants(perm.Admin))
	assert.True(t, perm.Admin.Grants(perm.Staff))
	assert.True(t, perm.Admin.Grants(perm.Basic))
	assert.True(t, !perm.Staff.Grants(perm.Admin))
	assert.True(t, perm.Staff.Grants(perm.Staff))
	assert.True(t, perm.Staff.Grants(perm.Basic))
	assert.True(t, !perm.Basic.Grants(perm.Admin))
	assert.True(t, !perm.Basic.Grants(perm.Staff))
	assert.True(t, perm.Basic.Grants(perm.Basic))
}
