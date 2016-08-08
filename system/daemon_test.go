package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rvolosatovs/systemgo/test/mock_unit"
	"github.com/rvolosatovs/systemgo/unit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUnit struct {
	*mock_unit.MockInterface
	*mock_unit.MockStarter
	*mock_unit.MockStopper
}

func newMock(ctrl *gomock.Controller) (u *mockUnit) {
	return &mockUnit{
		MockInterface: mock_unit.NewMockInterface(ctrl),
		MockStarter:   mock_unit.NewMockStarter(ctrl),
		MockStopper:   mock_unit.NewMockStopper(ctrl),
	}
}

func TestGet(t *testing.T) {
	sys := New()
	sys.SetPaths(os.TempDir())

	name := "foo.service"

	fpath := filepath.Join(os.TempDir(), name)

	file, err := os.Create(fpath)
	require.NoError(t, err, "os.Create(fpath)")
	defer file.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := newMock(ctrl)

	m.MockInterface.EXPECT().Define(gomock.Any()).Return(nil).Times(1)

	u, err := sys.Supervise(name, m)
	require.NoError(t, err)

	sys.units[fpath] = u

	for _, name := range []string{name, fpath} {
		ptr, err := sys.Get(name)
		require.NoError(t, err, name)
		assert.Equal(t, u, ptr, name)
	}
}

func TestSuported(t *testing.T) {
	for suffix, is := range supported {
		assert.Equal(t, is, Supported("foo"+suffix))
	}

	assert.False(t, Supported("foo.wrong"))
}

func TestPathset(t *testing.T) {
	path, err := ioutil.TempDir("", "pathset-test")
	require.NoError(t, err, "ioutil.TempDir")
	defer os.RemoveAll(path)

	cases := []string{
		"foo.service",
		"foo.mount",
		"foo.socket",
		"foo.target",
		"foo.wrong",
	}

	correct := 0
	for _, name := range cases {
		err = ioutil.WriteFile(filepath.Join(path, name), []byte{}, 0666)
		require.NoError(t, err, "ioutil.WriteFile")

		if Supported(name) {
			correct++
		}
	}

	paths, err := pathset(path)
	require.NoError(t, err, "pathset")
	assert.Len(t, paths, correct, "paths")
}

func TestStart(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	abc := map[string]*mockUnit{
		"a": newMock(ctrl),
		"b": newMock(ctrl),
		"c": newMock(ctrl),
	}

	abc["a"].MockInterface.EXPECT().Requires().Return([]string{"b", "c"}).Times(1)
	abc["a"].MockInterface.EXPECT().After().Return([]string{"b", "c"}).Times(1)

	abc["b"].MockInterface.EXPECT().Requires().Return([]string{"c"}).Times(1)
	abc["b"].MockInterface.EXPECT().After().Return([]string{"c"}).Times(1)

	empty(abc["a"], "wants", "before", "conflicts")
	empty(abc["b"], "wants", "before", "conflicts")
	empty(abc["c"], "wants", "before", "conflicts", "after", "requires")

	for mocks, seq := range map[*map[string]*mockUnit][]string{
		&abc: {"a", "b", "c"},
	} {
		sys := New()

		for name, m := range *mocks {
			u, err := sys.Supervise(name, m)
			require.NoError(t, err)

			u.loaded = unit.Loaded
		}

		sequence(*mocks, seq)

		assert.NoError(t, sys.Start("a"), "sys.Start("+"a"+")")

		time.Sleep(time.Second)
	}
}

func TestStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := newMock(ctrl)

	sys := New()

	m.MockInterface.EXPECT().Conflicts().Return([]string{}).Times(1)

	u, err := sys.Supervise("TestStop", m)
	u.loaded = unit.Loaded
	require.NoError(t, err)

	assert.Error(t, sys.Stop(u.Name()))
	assert.NoError(t, sys.Stop(u.Name()))
}

func sequence(units map[string]*mockUnit, names []string) *gomock.Call {
	switch len(names) {
	case 0:
		return nil
	case 1:
		fmt.Printf("<-%s\n", names[0])
		return units[names[0]].MockStarter.EXPECT().Start().Return(nil).Times(1)
	default:
		fmt.Printf("<-%s", names[0])
		return units[names[0]].MockStarter.EXPECT().Start().Return(nil).After(sequence(units, names[1:])).Times(1)
	}
}

func empty(m *mockUnit, methods ...string) {
	for _, method := range methods {
		emptyOne(m, method).Times(1)
	}
}

func emptyOne(m *mockUnit, method string) (c *gomock.Call) {
	exp := m.MockInterface.EXPECT()
	switch method {
	case "requires":
		c = exp.Requires()
	case "wants":
		c = exp.Wants()
	case "before":
		c = exp.Before()
	case "after":
		c = exp.After()
	case "wantedBy":
		c = exp.WantedBy()
	case "requiredBy":
		c = exp.RequiredBy()
	case "conflicts":
		c = exp.Conflicts()
	}
	return c.Return([]string{})
}
