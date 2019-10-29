package perm_test

import (
	"testing"

	"github.com/enpointe/activity/perm"
	"gotest.tools/assert"
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
	assert.Assert(t, perm.Admin.Grants(perm.Admin))
	assert.Assert(t, perm.Admin.Grants(perm.Staff))
	assert.Assert(t, perm.Admin.Grants(perm.Basic))
	assert.Assert(t, !perm.Staff.Grants(perm.Admin))
	assert.Assert(t, perm.Staff.Grants(perm.Staff))
	assert.Assert(t, perm.Staff.Grants(perm.Basic))
	assert.Assert(t, !perm.Basic.Grants(perm.Admin))
	assert.Assert(t, !perm.Basic.Grants(perm.Staff))
	assert.Assert(t, perm.Basic.Grants(perm.Basic))
}
