// Package config handles loading and validating vaultpipe configuration.
//
// Configuration can be provided via a YAML or JSON file and is supplemented
// (or overridden) by the following environment variables:
//
//	VAULT_ADDR      — Vault server address (e.g. http://127.0.0.1:8200)
//	VAULT_TOKEN     — Vault authentication token
//	VAULT_NAMESPACE — Vault Enterprise namespace (optional)
//
// Example YAML configuration:
//
//	vault:
//	  address: http://127.0.0.1:8200
//	  token: s.mytoken
//	secrets:
//	  - path: secret/data/myapp
//	    prefix: MYAPP_
//	inject:
//	  overwrite: false
package config
