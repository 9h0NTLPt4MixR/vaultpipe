package process

import (
	"os"
	"os/signal"
	"syscall"
)

// SignalForwarder listens for OS signals and forwards them to a target process.
type SignalForwarder struct {
	proc    *os.Process
	signals []os.Signal
	stopCh  chan struct{}
}

// NewSignalForwarder creates a SignalForwarder that will relay the given signals
// to proc. If no signals are provided, SIGINT, SIGTERM, and SIGHUP are used.
func NewSignalForwarder(proc *os.Process, sigs ...os.Signal) *SignalForwarder {
	if len(sigs) == 0 {
		sigs = []os.Signal{syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP}
	}
	return &SignalForwarder{
		proc:    proc,
		signals: sigs,
		stopCh:  make(chan struct{}),
	}
}

// Start begins forwarding signals in a background goroutine.
func (sf *SignalForwarder) Start() {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sf.signals...)

	go func() {
		defer signal.Stop(ch)
		for {
			select {
			case sig := <-ch:
				if sf.proc != nil {
					_ = sf.proc.Signal(sig)
				}
			case <-sf.stopCh:
				return
			}
		}
	}()
}

// Stop halts signal forwarding.
func (sf *SignalForwarder) Stop() {
	close(sf.stopCh)
}
