package audit

// SecretFetcher is satisfied by any type that can retrieve secrets by path.
type SecretFetcher interface {
	GetSecrets(path string) (map[string]string, error)
}

// InstrumentedFetcher wraps a SecretFetcher and emits audit events.
type InstrumentedFetcher struct {
	inner  SecretFetcher
	logger *Logger
}

// NewInstrumentedFetcher returns a SecretFetcher that audits every call.
func NewInstrumentedFetcher(f SecretFetcher, l *Logger) *InstrumentedFetcher {
	return &InstrumentedFetcher{inner: f, logger: l}
}

// GetSecrets delegates to the wrapped fetcher and logs the outcome.
func (i *InstrumentedFetcher) GetSecrets(path string) (map[string]string, error) {
	secrets, err := i.inner.GetSecrets(path)

	e := Event{
		Type:    EventSecretFetch,
		Path:    path,
		Success: err == nil,
	}
	if err != nil {
		e.Error = err.Error()
	} else {
		keys := make([]string, 0, len(secrets))
		for k := range secrets {
			keys = append(keys, k)
		}
		e.Keys = keys
	}
	_ = i.logger.Log(e) // best-effort; do not mask the original error

	return secrets, err
}
