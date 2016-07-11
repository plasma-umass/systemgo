package system

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/b1101/systemgo/test/mock_unit"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
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

func TestOrder(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
}
