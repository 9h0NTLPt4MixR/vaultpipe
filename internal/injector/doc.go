// Package injector provides utilities for injecting Vault secrets into
// process environments at runtime.
//
// The primary type is Injector, which accepts a map of secret key/value pairs
// retrieved from HashiCorp Vault and sets them as OS environment variables.
// Keys are automatically uppercased and can optionally be prefixed.
//
// Basic usage:
//
//	inj := injector.NewInjector(
//		injector.WithPrefix("myapp"),
//		injector.WithOverwrite(false),
//	)
//
//	if err := inj.Inject(secrets); err != nil {
//		log.Fatal(err)
//	}
//
// The EnvWriter type can be used to serialize injected variables to an
// io.Writer in .env file format, useful for debugging or writing dotenv files.
package injector
