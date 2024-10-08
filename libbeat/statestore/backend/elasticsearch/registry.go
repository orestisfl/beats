package elasticsearch

import (
	"errors"
	"runtime/debug"
	"sync"

	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/statestore/backend"
)

func New(log *logp.Logger, b *beat.Beat) backend.Registry {
	log.Infow("Orestis: Using elasticsearch as state store", "beat", b)
	return &registry{
		log:    log,
		beat:   b,
		stores: make(map[string]*store),
		lock:   sync.Mutex{},
	}
}

type registry struct {
	log    *logp.Logger
	beat   *beat.Beat
	stores map[string]*store
	lock   sync.Mutex
}

func (s *registry) Access(name string) (backend.Store, error) {
	l := s.log.With(
		"name",
		name,
		"beat",
		s.beat.Info,
		"BeatConfig",
		s.beat.BeatConfig,
		"InitialConfig",
		s.beat.Config,
	)
	l.Infow("Accessing elasticsearch store", "stack", debug.Stack())

	s.lock.Lock()
	defer s.lock.Unlock()

	if st, ok := s.stores[name]; ok {
		return st, nil
	}
	st := openStore(l, s.beat.Info, name)
	s.stores[name] = st
	return st, nil
}

func (s *registry) Close() error {
	s.log.Info("Closing elasticsearch state store registry")

	s.lock.Lock()
	defer s.lock.Unlock()

	var err error
	for _, st := range s.stores {
		err = errors.Join(err, st.Close())
	}
	return err
}
