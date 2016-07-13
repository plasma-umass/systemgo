package system

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/b1101/systemgo/test/mock_unit"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var deps = struct {
	defined, onDisk []string
}{
	[]string{"foo.service", "bar.service"},
	[]string{"test.service", "test.target"},
}

func init() {
	for i, name := range deps.onDisk {
		deps.onDisk[i] = filepath.Join(os.TempDir(), name)
	}
}

func TestDeps(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mock_unit.NewMockInterface(ctrl)
	m.EXPECT().Wants().Return(deps.defined).Times(1)
	m.EXPECT().Requires().Return(deps.defined).Times(1)

	u := NewUnit(m)
	u.path = filepath.Join(os.TempDir(), "testDeps")

	for _, suffix := range []string{".requires", ".wants"} {
		dirpath := u.path + suffix
		require.NoError(t, os.Mkdir(dirpath, 0755))
		defer os.RemoveAll(dirpath)

		for _, name := range append(deps.onDisk[:], "wrong.unittype", "foo.bar") {
			f, err := os.Create(name)
			defer os.Remove(name)

			require.NoError(t, err)
			require.NoError(t, f.Close())
			require.NoError(t, os.Symlink(name, filepath.Join(dirpath, filepath.Base(name))))
		}
	}

	expected := append(deps.defined[:], deps.onDisk[:]...)

	for _, deps := range [][]string{u.Wants(), u.Requires()} {
		for _, dep := range deps {
			assert.Contains(t, expected, dep)
		}
	}
}
