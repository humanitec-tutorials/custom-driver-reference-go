package config

import (
	"os"
	"path"
	"runtime"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	// As tests run from the current dir we need this trick to change dir as if they run from the module's root
	_, filename, _, _ := runtime.Caller(0)
	dir := path.Join(path.Dir(filename), "../..")
	os.Chdir(dir)

	code := m.Run()
	os.Exit(code)
}

func TestConfig(t *testing.T) {
	_, err := GetConfig()
	assert.NoError(t, err)
}

func TestConfigFromEnv(t *testing.T) {
	expectedFakeAWSClientEnabled := true
	expectedLogLevel := "info"
	t.Setenv("FAKE_AWS_CLIENT", strconv.FormatBool(expectedFakeAWSClientEnabled))
	t.Setenv("LOG_LEVEL", expectedLogLevel)

	conf, err := GetConfig()
	assert.NoError(t, err)

	assert.Equal(t, expectedFakeAWSClientEnabled, conf.FakeAWSClient)
	assert.Equal(t, expectedLogLevel, conf.LogLevel)
}
