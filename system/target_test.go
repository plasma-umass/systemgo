package system

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/plasma-umass/systemgo/test/mock_unit"
	"github.com/plasma-umass/systemgo/unit"
	"github.com/stretchr/testify/assert"
)

func TestTargetActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sys := New()
	targ := &Target{System: sys}

	for name, st := range map[string]unit.Activation{
		"active":       unit.Active,
		"inactive":     unit.Inactive,
		"reloading":    unit.Reloading,
		"failed":       unit.Failed,
		"activating":   unit.Activating,
		"deactivating": unit.Deactivating,
	} {
		u := mock_unit.NewMockInterface(ctrl)
		u.EXPECT().Active().Return(st).AnyTimes()
		sys.Supervise(name, u)
	}

	for deps, expected := range map[*[]string]unit.Activation{
		{"non-existent"}:                       unit.Inactive,
		{"active"}:                             unit.Active,
		{"active", "inactive"}:                 unit.Inactive,
		{"active", "inactive", "failed"}:       unit.Failed,
		{"active", "inactive", "activating"}:   unit.Activating,
		{"active", "inactive", "deactivating"}: unit.Deactivating,
		{"active", "inactive", "reloading"}:    unit.Reloading,
	} {
		targ.Definition.Unit.Requires = *deps
		assert.Equal(t, expected, targ.Active(), fmt.Sprintf("Deps: %v", *deps))
	}
}
