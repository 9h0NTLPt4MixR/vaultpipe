# vaultpipe

Lightweight secrets injection middleware that pulls from HashiCorp Vault and injects into process environments at runtime.

---

## Installation

```bash
go install github.com/yourusername/vaultpipe@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultpipe.git && cd vaultpipe && go build ./...
```

---

## Usage

`vaultpipe` wraps your process and injects secrets as environment variables before execution.

```bash
vaultpipe --addr=https://vault.example.com \
           --token=$VAULT_TOKEN \
           --secret=secret/data/myapp \
           -- ./myapp --start
```

You can also use a config file:

```yaml
# vaultpipe.yaml
vault:
  addr: https://vault.example.com
  token_env: VAULT_TOKEN
secrets:
  - path: secret/data/myapp
    env_prefix: APP_
```

```bash
vaultpipe --config=vaultpipe.yaml -- ./myapp
```

Secrets stored at the given Vault path are mapped to environment variables and made available to the child process at runtime. The parent process retains no secret values after handoff.

---

## Environment Variables

| Variable | Description |
|---|---|
| `VAULT_ADDR` | Vault server address |
| `VAULT_TOKEN` | Authentication token |
| `VAULTPIPE_CONFIG` | Path to config file |

---

## Contributing

Pull requests are welcome. Please open an issue first to discuss proposed changes.

---

## License

MIT © 2024 yourusername