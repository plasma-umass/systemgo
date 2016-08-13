package system

import (
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rvolosatovs/systemgo/system"
	"github.com/rvolosatovs/systemgo/test/mock_unit"
	"github.com/rvolosatovs/systemgo/unit"
	"github.com/stretchr/testify/assert"
)

func TestActive(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	//targ := &Target{System:New()}
	targ := &Target{Get: func(name string) (unit.Subber, error) {
		u := mock_unit.NewMockInterface(ctrl)
		switch name {
		case "active":
			u.EXPECT().Active().Return(unit.Active).Times(1)
		case "inactive":
			u.EXPECT().Active().Return(unit.Inactive).Times(1)
		case "reloading":
			u.EXPECT().Active().Return(unit.Reloading).Times(1)
		case "failed":
			u.EXPECT().Active().Return(unit.Failed).Times(1)
		case "activating":
			u.EXPECT().Active().Return(unit.Activating).Times(1)
		case "deactivating":
			u.EXPECT().Active().Return(unit.Deactivating).Times(1)
		default:
			return nil, system.ErrNotFound
		}
		return u, nil
	}}

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
		assert.Equal(t, targ.Active(), expected, fmt.Sprintf("Should be %s, not %s\ndeps: %v", expected, targ.Active(), *deps))
	}
}

func TestStart(t *testing.T) {
	targ := &Target{}
	assert.Equal(t, targ.Start(), nil)
}

func TestStop(t *testing.T) {
	targ := &Target{}
	assert.Equal(t, targ.Stop(), nil)
}
