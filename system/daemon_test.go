package system

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/b1101/systemgo/test/mock_unit"
	"github.com/b1101/systemgo/unit"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	uInt := mock_unit.NewMockInterface(ctrl)
	uInt.EXPECT().Define(gomock.Any()).Return(nil).Times(2)

	u := NewUnit(uInt)

	sys.parsed[name] = u
	sys.parsed[fpath] = u

	for _, name := range []string{name, fpath} {
		ptr, err := sys.Get(name)
		require.NoError(t, err, "sys.Get")
		assert.Equal(t, u, ptr, "*Unit")
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

	abc := map[string]*mock_unit.MockInterface{
		"a": mock_unit.NewMockInterface(ctrl),
		"b": mock_unit.NewMockInterface(ctrl),
		"c": mock_unit.NewMockInterface(ctrl),
	}

	abc["a"].EXPECT().Requires().Return([]string{"b", "c"}).Times(2)
	abc["b"].EXPECT().Requires().Return([]string{"c"}).Times(2)

	abc["a"].EXPECT().After().Return([]string{"b", "c"}).Times(1)
	abc["b"].EXPECT().After().Return([]string{"c"}).Times(1)

	abc["b"].EXPECT().Active().Return(unit.Active).Times(1)
	abc["c"].EXPECT().Active().Return(unit.Active).Times(2)

	empty(abc["a"], "wants", "before", "conflicts")
	empty(abc["b"], "wants", "before", "conflicts")
	empty(abc["c"], "wants", "after", "before", "conflicts", "requires")

	for mocks, seq := range map[*map[string]*mock_unit.MockInterface][]string{
		&abc: {"a", "b", "c"},
	} {
		sys := New()

		for name, m := range *mocks {
			u := NewUnit(m)
			sys.loaded[name] = u
			sys.names[u] = name
		}

		fmt.Println("----- Units -----")
		fmt.Println("pointer\t\tname")
		for name, u := range sys.loaded {
			fmt.Printf("%p\t%s\n", u, name)
		}
		fmt.Println("-----------------")

		sequence(*mocks, seq)

		assert.NoError(t, sys.Start(seq[0]), "sys.Start("+seq[0]+")")

		time.Sleep(time.Second)
	}
}

func sequence(units map[string]*mock_unit.MockInterface, names []string) *gomock.Call {
	switch len(names) {
	case 0:
		return nil
	case 1:
		fmt.Printf("<-%s\n", names[0])
		return units[names[0]].EXPECT().Start().Return(nil).Times(1)
	default:
		fmt.Printf("<-%s", names[0])
		return units[names[0]].EXPECT().Start().Return(nil).After(sequence(units, names[1:])).Times(1)
	}
}

func empty(m *mock_unit.MockInterface, methods ...string) {
	for _, method := range methods {
		c := emptyOne(m, method)
		if method == "requires" {
			c.Times(2)
		} else {
			c.Times(1)
		}
	}
}

func emptyOne(m *mock_unit.MockInterface, method string) (c *gomock.Call) {
	exp := m.EXPECT()
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
