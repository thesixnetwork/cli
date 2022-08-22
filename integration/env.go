package envtest

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/cenkalti/backoff"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/ignite/cli/ignite/chainconfig"
	"github.com/ignite/cli/ignite/pkg/cmdrunner"
	"github.com/ignite/cli/ignite/pkg/cmdrunner/step"
	"github.com/ignite/cli/ignite/pkg/cosmosfaucet"
	"github.com/ignite/cli/ignite/pkg/httpstatuschecker"
	"github.com/ignite/cli/ignite/pkg/xexec"
	"github.com/ignite/cli/ignite/pkg/xurl"
)

const (
	IgniteApp = "ignite"
	Stargate  = "stargate"
)

var isCI, _ = strconv.ParseBool(os.Getenv("CI"))

// ConfigUpdateFunc defines a function type to update config file values.
type ConfigUpdateFunc func(*chainconfig.Config) error

// Env provides an isolated testing environment and what's needed to
// make it possible.
type Env struct {
	t   *testing.T
	ctx context.Context
}

// New creates a new testing environment.
func New(t *testing.T) Env {
	ctx, cancel := context.WithCancel(context.Background())
	e := Env{
		t:   t,
		ctx: ctx,
	}
	t.Cleanup(cancel)

	if !xexec.IsCommandAvailable(IgniteApp) {
		t.Fatal("ignite needs to be installed")
	}

	return e
}

// SetCleanup registers a function to be called when the test (or subtest) and all its
// subtests complete.
func (e Env) SetCleanup(f func()) {
	e.t.Cleanup(f)
}

// Ctx returns parent context for the test suite to use for cancelations.
func (e Env) Ctx() context.Context {
	return e.ctx
}

type clientOptions struct {
	env                    map[string]string
	testName, testFilePath string
}

// ClientOption defines options for the TS client test runner.
type ClientOption func(*clientOptions)

// ClientEnv option defines environment values for the tests.
func ClientEnv(env map[string]string) ClientOption {
	return func(o *clientOptions) {
		for k, v := range env {
			o.env[k] = v
		}
	}
}

// ClientTestName option defines a pattern to match the test names that should be run.
func ClientTestName(pattern string) ClientOption {
	return func(o *clientOptions) {
		o.testName = pattern
	}
}

// ClientTestFile option defines the name of the file where to look for tests.
func ClientTestFile(filePath string) ClientOption {
	return func(o *clientOptions) {
		o.testFilePath = filePath
	}
}

// IsAppServed checks that app is served properly and servers are started to listening
// before ctx canceled.
func (e Env) IsAppServed(ctx context.Context, host chainconfig.Host) error {
	checkAlive := func() error {
		addr, err := xurl.HTTP(host.API)
		if err != nil {
			return err
		}

		ok, err := httpstatuschecker.Check(ctx, fmt.Sprintf("%s/cosmos/base/tendermint/v1beta1/node_info", addr))
		if err == nil && !ok {
			err = errors.New("app is not online")
		}
		if HasTestVerboseFlag() {
			fmt.Printf("IsAppServed at %s: %v\n", addr, err)
		}
		return err
	}

	return backoff.Retry(checkAlive, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}

// IsFaucetServed checks that faucet of the app is served properly
func (e Env) IsFaucetServed(ctx context.Context, faucetClient cosmosfaucet.HTTPClient) error {
	checkAlive := func() error {
		_, err := faucetClient.FaucetInfo(ctx)
		return err
	}

	return backoff.Retry(checkAlive, backoff.WithContext(backoff.NewConstantBackOff(time.Second), ctx))
}

// TmpDir creates a new temporary directory.
func (e Env) TmpDir() (path string) {
	return e.t.TempDir()
}

// Home returns user's home dir.
func (e Env) Home() string {
	home, err := os.UserHomeDir()
	require.NoError(e.t, err)
	return home
}

// UpdateConfig updates config.yml file values.
func (e Env) UpdateConfig(path, configFile string, update ConfigUpdateFunc) {
	if configFile == "" {
		configFile = ConfigYML
	}

	f, err := os.OpenFile(filepath.Join(path, configFile), os.O_RDWR|os.O_CREATE, 0o755)
	require.NoError(e.t, err)

	defer f.Close()

	var cfg chainconfig.Config

	require.NoError(e.t, yaml.NewDecoder(f).Decode(&cfg))
	require.NoError(e.t, update(&cfg))
	require.NoError(e.t, f.Truncate(0))

	_, err = f.Seek(0, 0)

	require.NoError(e.t, err)
	require.NoError(e.t, yaml.NewEncoder(f).Encode(cfg))
}

// AppHome returns app's root home/data dir path.
func (e Env) AppHome(name string) string {
	return filepath.Join(e.Home(), fmt.Sprintf(".%s", name))
}

// RunClientTests runs the Typescript client tests.
func (e Env) RunClientTests(path string, options ...ClientOption) bool {
	npm, err := exec.LookPath("npm")
	require.NoError(e.t, err, "npm binary not found")

	// The root dir for the tests must be an absolute path.
	// It is used as the start search point to find test files.
	rootDir, err := os.Getwd()
	require.NoError(e.t, err)

	// The filename of this module is required to be able to define the location
	// of the TS client test runner package to be used as working directory when
	// running the tests.
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		e.t.Fatal("failed to read file name")
	}

	opts := clientOptions{
		env: map[string]string{
			"TEST_CHAIN_PATH": path,
		},
	}
	for _, o := range options {
		o(&opts)
	}

	var (
		output bytes.Buffer
		env    []string
	)

	//  Install the dependencies needed to run TS client tests
	ok = e.Exec("install client dependencies", step.NewSteps(
		step.New(
			step.Workdir(fmt.Sprintf("%s/vue", path)),
			step.Stdout(&output),
			step.Exec(npm, "install"),
			step.PostExec(func(err error) error {
				// Print the npm output when there is an error
				if err != nil {
					e.t.Log("\n", output.String())
				}

				return err
			}),
		),
	))
	if !ok {
		return false
	}

	output.Reset()

	args := []string{"run", "test", "--", "--dir", rootDir}
	if opts.testName != "" {
		args = append(args, "-t", opts.testName)
	}

	if opts.testFilePath != "" {
		args = append(args, opts.testFilePath)
	}

	for k, v := range opts.env {
		env = append(env, cmdrunner.Env(k, v))
	}

	// The tests are run from the TS client test runner package directory
	runnerDir := filepath.Join(filepath.Dir(filename), "testdata/tstestrunner")

	// TODO: Ignore stderr ? Errors are already displayed with traceback in the stdout
	return e.Exec("run client tests", step.NewSteps(
		// Make sure the test runner dependencies are installed
		step.New(
			step.Workdir(runnerDir),
			step.Stdout(&output),
			step.Exec(npm, "install"),
			step.PostExec(func(err error) error {
				// Print the npm output when there is an error
				if err != nil {
					e.t.Log("\n", output.String())
				}

				return err
			}),
		),
		// Run the TS client tests
		step.New(
			step.Workdir(runnerDir),
			step.Stdout(&output),
			step.Env(env...),
			step.PreExec(func() error {
				// Clear the output from the previous step
				output.Reset()

				return nil
			}),
			step.Exec(npm, args...),
			step.PostExec(func(err error) error {
				// Always print tests output to be available on errors or when verbose is enabled
				e.t.Log("\n", output.String())

				return err
			}),
		),
	))
}

// Must fails the immediately if not ok.
// t.Fail() needs to be called for the failing tests before running Must().
func (e Env) Must(ok bool) {
	if !ok {
		e.t.FailNow()
	}
}

func (e Env) HasFailed() bool {
	return e.t.Failed()
}

func (e Env) RequireExpectations() {
	e.Must(e.HasFailed())
}

func Contains(s, partial string) bool {
	return strings.Contains(s, strings.TrimSpace(partial))
}

func HasTestVerboseFlag() bool {
	return flag.Lookup("test.v").Value.String() == "true"
}
