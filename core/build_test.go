package core

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockIntegration is an autogenerated mock type for the MockIntegration type
type MockIntegration struct {
	mock.Mock
}

// AttachToApp provides a mock function with given fields: _a0
func (_m *MockIntegration) AttachToApp(_a0 App) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(App) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Identifier provides a mock function with given fields:
func (_m *MockIntegration) Identifier() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// IsProvider provides a mock function with given fields: _a0
func (_m *MockIntegration) IsProvider(_a0 string) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// ProvideFor provides a mock function with given fields: c, directory
func (_m *MockIntegration) ProvideFor(c *BuildConfig, directory string) error {
	ret := _m.Called(c, directory)

	var r0 error
	if rf, ok := ret.Get(0).(func(*BuildConfig, string) error); ok {
		r0 = rf(c, directory)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Shutdown provides a mock function with given fields:
func (_m *MockIntegration) Shutdown() {
	_m.Called()
}

func getSuccessfulIntegration() *MockIntegration {
	i := &MockIntegration{}
	i.On("Identifier").Return("Success")
	i.On("IsProvider", mock.Anything).Return(true)
	i.On("ProvideFor", mock.AnythingOfType("*core.BuildConfig"), mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
		dir := args.Get(1).(string)
		//FIXME - this is lazy, stops tests running on windows, is bad in general, i'm so tired
		cmd := exec.Command("cp", "testdata/failure.sh", "testdata/success.sh", "testdata/fiveminutes.sh", dir)
		cmd.Run()
	}).Return(nil)
	return i
}

func getFailedIntegration() *MockIntegration {
	i := &MockIntegration{}
	i.On("Identifier").Return("Success")
	i.On("IsProvider", mock.Anything).Return(true)
	i.On("ProvideFor", mock.Anything, mock.Anything).Return(errors.New("testmarker"))
	return i
}

func TestProvisionBuildIntoDirectory(t *testing.T) {
	assert := assert.New(t)
	require := require.New(t)

	b := build{token: "testtoken"}
	dir, err := provisionDirectory()
	require.NoError(err)

	integrationSuccess := getSuccessfulIntegration()
	integrationFailure := getFailedIntegration()

	config := BuildConfig{
		BaseRepo:     "testmarker-baserepo",
		MergeRepo:    "testmarker-mergerepo",
		Integrations: []Integration{integrationFailure, integrationSuccess},
	}

	assert.NoError(b.provisionBuildIntoDirectory(&config, dir))
	assert.NoError(cleanupDirectory(dir))
}

func TestRunBuildSync(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	b := build{token: "testtoken"}
	b.config = &BuildConfig{
		Integrations: []Integration{getSuccessfulIntegration()},
		BuildRunner:  "success.sh",
		Deadline:     time.Second * 5,
	}
	b.Ref()

	require.NoError(b.runBuildSync(*b.config))

	stdoutpipe, err := b.Stdout()
	require.NoError(err)
	stdout, err := ioutil.ReadAll(stdoutpipe)
	require.NoError(err)
	assert.EqualValues([]byte("testmarker\n"), stdout)

	stderrpipe, err := b.Stderr()
	require.NoError(err)
	stderr, err := ioutil.ReadAll(stderrpipe)
	require.NoError(err)
	assert.EqualValues([]byte("testmarker\n"), stderr)
	b.Unref()
	assert.Empty(b.buildDirectory)

	require.True(b.state.HasStopped())
}

func TestRunBuildSyncFailure(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	b := build{token: "testtoken"}
	b.config = &BuildConfig{
		Integrations: []Integration{getSuccessfulIntegration()},
		BuildRunner:  "failure.sh",
		Deadline:     time.Second * 5,
	}
	b.Ref()

	require.Error(b.runBuildSync(*b.config))

	stdoutpipe, err := b.Stdout()
	require.NoError(err)
	stdout, err := ioutil.ReadAll(stdoutpipe)
	require.NoError(err)
	assert.EqualValues([]byte("testmarker\n"), stdout)

	stderrpipe, err := b.Stderr()
	require.NoError(err)
	stderr, err := ioutil.ReadAll(stderrpipe)
	require.NoError(err)
	assert.EqualValues([]byte("testmarker\n"), stderr)
	b.Unref()
	assert.Empty(b.buildDirectory)
	require.True(b.state.HasStopped())
}

func TestRunBuildSyncDeadline(t *testing.T) {
	require := require.New(t)
	assert := assert.New(t)

	b := build{token: "testtoken"}
	b.config = &BuildConfig{
		Integrations: []Integration{getSuccessfulIntegration()},
		BuildRunner:  "fiveminutes.sh",
		Deadline:     time.Second,
	}
	b.Ref()

	require.Error(b.runBuildSync(*b.config))

	stdoutpipe, err := b.Stdout()
	require.NoError(err)
	stdout, err := ioutil.ReadAll(stdoutpipe)
	require.NoError(err)
	assert.EqualValues([]byte("testmarker\n"), stdout)
	b.Unref()
	assert.Empty(b.buildDirectory)
	require.True(b.state.HasStopped())
}