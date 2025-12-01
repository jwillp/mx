package mxtest

import (
	"github.com/morebec/misas/misas"
	"github.com/samber/lo"
)

type MockCommand struct{ tn misas.CommandTypeName }

func (m MockCommand) TypeName() misas.CommandTypeName {
	return lo.Ternary(m.tn != "", m.tn, "mxtest.MockCommand")
}
