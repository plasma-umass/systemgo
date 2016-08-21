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
	"github.com/plasma-umass/systemgo/test/mock_unit"
	"github.com/plasma-umass/systemgo/unit"
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

	for name, mock := range mocks {
		mock.MockInterface.EXPECT().Active().Return(unit.Inactive).AnyTimes()

		u, err := sys.Supervise(name, mock)
		require.NoError(t, err)

		u.load = unit.Loaded
	}

	calls := make([]*gomock.Call, len(sequence))
	for i, name := range sequence {
		calls[i] = mocks[name].MockStarter.EXPECT().Start().Return(nil).Times(1)
	}
	gomock.InOrder(calls...)

	require.NoError(t, sys.Start("c"), "sys.Start("+"c"+")")

	names := make([]string, 0, len(mocks))
	for name := range mocks {
		names = append(names, name)
	}
	waitForJobs(t, sys, names...)
}

func TestStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := newMock(ctrl)
	m.MockStopper.EXPECT().Stop().Return(nil).Times(1)
	m.MockInterface.EXPECT().Active().Return(unit.Active).AnyTimes()

	sys := New()

	u, err := sys.Supervise("TestStop", m)
	u.load = unit.Loaded
	require.NoError(t, err)

	require.NoError(t, sys.Stop("TestStop"))
	waitForJobs(t, sys, "TestStop")
}

func TestIsolate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sys := New()

	mocks := map[string]*mockUnit{
		"a": newMock(ctrl),
		"b": newMock(ctrl),
		"c": newMock(ctrl),
	}

	mocks["a"].MockStopper.EXPECT().Stop().Return(nil).Times(1)
	mocks["b"].MockStopper.EXPECT().Stop().Return(nil).Times(1)

	empty(mocks["c"], "wants", "before", "conflicts", "after", "requires")

	for name, mock := range mocks {
		mock.MockInterface.EXPECT().Active().Return(unit.Active).AnyTimes()

		u, err := sys.Supervise(name, mock)
		require.NoError(t, err)

		u.load = unit.Loaded
	}

	require.NoError(t, sys.Isolate("c"), "sys.Isolate")

	names := make([]string, 0, len(mocks))
	for name := range mocks {
		names = append(names, name)
	}
	waitForJobs(t, sys, "a", "b")
}

func waitForJobs(t *testing.T, sys *Daemon, names ...string) {
	wg := &sync.WaitGroup{}
	for _, name := range names {
		u, err := sys.Unit(name)
		require.NoError(t, err, "sys.Unit", name)

		wg.Add(1)
		go func(name string, u *Unit) {
			defer wg.Done()

			for u.job == nil {
				log.Warnf("%s job still nil", name)
				time.Sleep(100 * time.Millisecond)
			}

			log.Warnf("Waiting for %s job to finish", name)
			u.job.Wait()

			assert.True(t, u.job.Success())
		}(name, u)
	}
	wg.Wait()
}

func TestEnable(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	sys := New()

	m := mock_unit.NewMockInterface(ctrl)
	m.EXPECT().WantedBy().Return([]string{"test.target"}).Times(2)
	m.EXPECT().RequiredBy().Return([]string{"test.target"}).Times(2)

	var err error

	for name, iface := range map[string]unit.Interface{
		"test.target":  nil,
		"test.service": m,
	} {
		e := log.WithField("name", name)

		var u *Unit
		if u, err = sys.Supervise(name, iface); err != nil {
			e.WithField("err", err).Fatal("sys.Supervise")
		}

		var f *os.File
		if f, err = ioutil.TempFile("", name); err != nil {
			e.WithField("err", err).Fatal("ioutil.TempDir")
		}

		u.path = f.Name()
		u.load = unit.Loaded
	}

	require.NoError(t, sys.Enable("test.service"), "sys.Enable")

	for _, suffix := range []string{"wants", "requires"} {
		path, err := os.Readlink(filepath.Join(sys.units["test.target"].path+"."+suffix, "test.service"))
		require.NoError(t, err, "os.Readlink")
		assert.Equal(t, path, sys.units["test.service"].path, "link path")
	}

	// TODO implement
	//st, err := sys.IsEnabled("test.service")
	//assert.NoError(t, err, "sys.IsEnabled")
	//assert.Equal(t, unit.Enabled, st, "sys.IsEnabled")

	require.NoError(t, sys.Disable("test.service"), "sys.Disable")
	for _, suffix := range []string{"wants", "requires"} {
		_, err := os.Open(filepath.Join(sys.units["test.target"].path+"."+suffix, "test.service"))
		assert.True(t, os.IsNotExist(err), "os.Open")
	}

	// TODO implement
	//st, err = sys.IsEnabled("test.service")
	//assert.NoError(t, err, "sys.IsEnabled")
	//assert.Equal(t, unit.Disabled, st, "sys.IsEnabled")
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
