package system

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
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
	defer log.SetLevel(log.GetLevel())
	log.SetLevel(log.DebugLevel)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mocks := map[string]*mockUnit{
		"a": newMock(ctrl),
		"b": newMock(ctrl),
		"c": newMock(ctrl),
	}

	sequence := []string{
		"a", "b", "c",
	}

	empty(mocks["a"], "wants", "before", "conflicts", "after", "requires")
	empty(mocks["b"], "wants", "before", "conflicts")
	empty(mocks["c"], "wants", "before", "conflicts")

	mocks["b"].MockInterface.EXPECT().After().Return([]string{"a"}).Times(1)
	mocks["b"].MockInterface.EXPECT().Requires().Return([]string{"a"}).Times(1)

	mocks["c"].MockInterface.EXPECT().After().Return([]string{"b", "a"}).Times(1)
	mocks["c"].MockInterface.EXPECT().Requires().Return([]string{"b", "a"}).Times(1)

	sys := New()

	units := map[string]*Unit{}
	for name, m := range mocks {
		u, err := sys.Supervise(name, m)
		require.NoError(t, err)

		u.loaded = unit.Loaded

		units[name] = u
	}

	calls := make([]*gomock.Call, len(sequence))
	for i, name := range sequence {
		calls[i] = mocks[name].MockStarter.EXPECT().Start().Return(nil).Times(1)
	}
	gomock.InOrder(calls...)

	assert.NoError(t, sys.Start("c"), "sys.Start("+"c"+")")

	wg := &sync.WaitGroup{}
	for _, u := range units {
		wg.Add(1)
		go func(u *Unit) {
			defer wg.Done()

			for u.job == nil {
				log.Warnf("%s job still nil", u.Name())
				time.Sleep(100 * time.Millisecond)
			}

			log.Infof("Waiting for %s job to finish", u.Name())
			u.job.Wait()
			log.Debugf("%s job finished", u.Name())

			assert.True(t, u.job.Success())
		}(u)
	}
	wg.Wait()
}

func TestStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := newMock(ctrl)
	m.MockStopper.EXPECT().Stop().Return(nil).Times(1)

	sys := New()

	u, err := sys.Supervise("TestStop", m)
	u.loaded = unit.Loaded
	require.NoError(t, err)

	assert.NoError(t, sys.Stop(u.Name()))
	for u.job == nil {
		log.Warnf("%s job still nil", u.Name())
		time.Sleep(100 * time.Millisecond)
	}
	u.job.Wait()

	assert.True(t, u.job.Success())
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
