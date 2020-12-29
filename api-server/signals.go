package apiserver

import (
	"sync"
)

// var (
// 	log = kubelog.Log.WithName("api-server")
// )

type errSignaler struct {
	// errSignal indicates that an error occurred, when closed.  It shouldn't
	// be written to.
	errSignal chan struct{}
	// err is the received error
	err error
	mu  sync.Mutex
}

func (r *errSignaler) SignalError(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err == nil {
		// non-error, ignore
		log.Error(nil, "SignalError called without an (with a nil) error, which should never happen, ignoring")
		return
	}
	if r.err != nil {
		// we already have an error, don't try again
		return
	}
	// save the error and report it
	r.err = err
	close(r.errSignal)
}

func (r *errSignaler) Error() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.err
}

func (r *errSignaler) GotError() chan struct{} {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.errSignal
}
