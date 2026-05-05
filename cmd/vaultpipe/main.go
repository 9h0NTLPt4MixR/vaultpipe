// Package main is the entry point for the vaultpipe CLI.
// It wires together configuration loading, Vault client initialisation,
// secret injection, and child-process execution.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yourusername/vaultpipe/internal/cache"
	"github.com/yourusername/vaultpipe/internal/config"
	"github.com/yourusername/vaultpipe/internal/injector"
	"github.com/yourusername/vaultpipe/internal/logger"
	"github.com/yourusername/vaultpipe/internal/process"
	"github.com/yourusername/vaultpipe/internal/vault"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "vaultpipe: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration from file and environment.
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	// Initialise structured logger.
	log := logger.New(os.Stderr, cfg.LogLevel)

	// Require a command to exec.
	args := os.Args[1:]
	if len(args) == 0 {
		return fmt.Errorf("usage: vaultpipe <command> [args...]")
	}

	// Build Vault client.
	vc, err := vault.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("creating vault client: %w", err)
	}

	// Set up an in-memory cache for secrets.
	secretCache := cache.New(cache.WithTTL(cfg.CacheTTL))

	// Fetch secrets, using the cache when available.
	secrets, err := fetchSecrets(context.Background(), vc, secretCache, cfg.SecretPaths, log)
	if err != nil {
		return fmt.Errorf("fetching secrets: %w", err)
	}

	// Build injector with configured options.
	injectOpts := []injector.Option{
		injector.WithOverwrite(cfg.Overwrite),
	}
	if cfg.EnvPrefix != "" {
		injectOpts = append(injectOpts, injector.WithPrefix(cfg.EnvPrefix))
	}
	inj := injector.NewInjector(injectOpts...)

	// Inject secrets into the current process environment.
	if err := inj.Inject(secrets); err != nil {
		return fmt.Errorf("injecting secrets: %w", err)
	}
	log.Info("secrets injected", "count", len(secrets))

	// Build the child-process runner.
	runner := process.NewRunner(
		process.WithEnv(os.Environ()),
		process.WithStdio(os.Stdin, os.Stdout, os.Stderr),
	)

	// Forward OS signals to the child process.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	forwarder := process.NewSignalForwarder(sigCh)
	defer forwarder.Stop()

	log.Info("starting child process", "command", args[0], "args", args[1:])
	if err := runner.Run(context.Background(), args); err != nil {
		return fmt.Errorf("child process: %w", err)
	}

	return nil
}

// fetchSecrets retrieves secrets for each configured path, consulting the
// cache before hitting Vault.
func fetchSecrets(
	ctx context.Context,
	vc *vault.Client,
	c *cache.Cache,
	paths []string,
	log *logger.Logger,
) (map[string]string, error) {
	allSecrets := make(map[string]string)

	for _, path := range paths {
		if cached, ok := c.Get(path); ok {
			log.Debug("cache hit", "path", path)
			for k, v := range cached {
				allSecrets[k] = v
			}
			continue
		}

		log.Debug("fetching from vault", "path", path)
		secrets, err := vc.GetSecrets(ctx, path)
		if err != nil {
			return nil, fmt.Errorf("path %q: %w", path, err)
		}

		c.Set(path, secrets)
		for k, v := range secrets {
			allSecrets[k] = v
		}
	}

	return allSecrets, nil
}
