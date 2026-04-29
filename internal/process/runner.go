// Package process provides utilities for spawning child processes
// with injected environment variables sourced from Vault.
package process

import (
	"errors"
	"os"
	"os/exec"
)

// Runner executes a subprocess with a merged environment.
type Runner struct {
	env     []string
	stdin   *os.File
	stdout  *os.File
	stderr  *os.File
}

// Option configures a Runner.
type Option func(*Runner)

// WithEnv sets additional environment variables (KEY=VALUE pairs) for the
// child process. These are merged with the current process environment.
func WithEnv(env []string) Option {
	return func(r *Runner) {
		r.env = append(r.env, env...)
	}
}

// WithStdio overrides the stdio streams for the child process.
func WithStdio(stdin, stdout, stderr *os.File) Option {
	return func(r *Runner) {
		r.stdin = stdin
		r.stdout = stdout
		r.stderr = stderr
	}
}

// NewRunner constructs a Runner with the provided options.
func NewRunner(opts ...Option) *Runner {
	r := &Runner{
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
	for _, o := range opts {
		o(r)
	}
	return r
}

// Run executes the given command and arguments as a child process.
// The child inherits the current environment merged with any extra env vars.
func (r *Runner) Run(name string, args ...string) error {
	if name == "" {
		return errors.New("process: command name must not be empty")
	}

	cmd := exec.Command(name, args...)
	cmd.Env = append(os.Environ(), r.env...)
	cmd.Stdin = r.stdin
	cmd.Stdout = r.stdout
	cmd.Stderr = r.stderr

	return cmd.Run()
}
