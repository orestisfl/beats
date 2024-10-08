package elasticsearch

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	conf "github.com/elastic/elastic-agent-libs/config"
	"github.com/elastic/elastic-agent-libs/logp"

	"github.com/elastic/beats/v7/libbeat/beat"
	"github.com/elastic/beats/v7/libbeat/common/transform/typeconv"
	"github.com/elastic/beats/v7/libbeat/esleg/eslegclient"
	"github.com/elastic/beats/v7/libbeat/statestore/backend"
	"github.com/elastic/beats/v7/x-pack/filebeat/tmp"
)

type store struct {
	log      *logp.Logger
	es       *eslegclient.Connection
	index    string
	ready    chan struct{}
	beatInfo beat.Info
	lock     sync.RWMutex
}

const docType = "_doc"

func openStore(log *logp.Logger, beatInfo beat.Info, name string) *store {
	log.Infow("Opening store", "name", name, "beat", beatInfo)
	return &store{
		log:      log,
		ready:    make(chan struct{}),
		index:    name,
		beatInfo: beatInfo,
		lock:     sync.RWMutex{},
	}
}

var (
	outputConf *conf.C
	lock       sync.RWMutex
)

func SetConfig(c *conf.C) {
	if c == nil {
		return
	}

	tmp.Debug("Setting elasticsearch state store configuration", "config", c)
	lock.Lock()
	defer lock.Unlock()
	outputConf = c
}

func getOutputConfig() *conf.C {
	for i := 0; i < 3; i++ {
		c := func() *conf.C {
			lock.RLock()
			defer lock.RUnlock()
			return outputConf
		}()
		if c != nil {
			tmp.Debug("Using elasticsearch state store configuration")
			return c
		}
		tmp.Debug("Waiting for elasticsearch state store configuration")
		time.Sleep(time.Second)
	}
	tmp.Debug("Failed to get elasticsearch state store configuration")
	return nil
}

func (s *store) Close() error {
	s.log.Infow("Close()")
	if s.es != nil {
		return s.es.Close()
	}
	return nil
}

func (s *store) Has(key string) (bool, error) {
	s.log.Infow("Has()", "key", key)
	if err := s.ensureEs(); err != nil {
		return false, err
	}

	sr, err := s.get(key)
	if err != nil {
		return false, err
	}
	return sr.Hits.Total.Value > 0, nil
}

func (s *store) Get(key string, value any) error {
	s.log.Infow("Get()", "key", key)
	tmp.Debug("Get()", "value", value)

	sr, err := s.get(key)
	if err != nil {
		return err
	}
	if sr.Hits.Total.Value == 0 {
		return errors.New("key not found")
	}
	return json.Unmarshal(sr.Hits.Hits[0], value)
}
func (s *store) get(key string) (*eslegclient.SearchResults, error) {
	if err := s.ensureEs(); err != nil {
		return nil, err
	}

	_, sr, err := s.es.SearchURIWithBody(s.index, docType, nil, map[string]any{
		"query": map[string]any{
			"term": map[string]string{
				"_id": key,
			},
		},
	})
	return sr, err
}

func (s *store) Set(key string, value any) error {
	s.log.Infow("Set()", "key", key)
	if err := s.ensureEs(); err != nil {
		return err
	}

	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, _, err = s.es.Index(s.index, docType, key, nil, jsonValue)
	return err
}

func (s *store) Remove(key string) error {
	s.log.Infow("Remove()", "key", key)
	if err := s.ensureEs(); err != nil {
		return err
	}

	_, _, err := s.es.Delete(s.index, docType, key, nil)
	return err
}

type entry struct {
	value map[string]interface{}
}

func (e entry) Decode(to interface{}) error {
	return typeconv.Convert(to, e.value)
}

func (s *store) Each(fn func(string, backend.ValueDecoder) (bool, error)) error {
	defer func() {
		recoverErr := recover()
		if recoverErr != nil {
			s.log.Errorw("panic in Each()", "error", recoverErr)
		}
	}()

	tmp.DebugWithStack("Each()", "index", s.index)
	if err := s.ensureEs(); err != nil {
		s.log.Warnw("failed to ensure ES", "error", err)
		return nil
	}

	_, sr, err := s.es.SearchURIWithBody(s.index, docType, nil, map[string]any{
		"query": map[string]any{
			"term": map[string]any{
				"match_all": struct{}{},
			},
		},
	})
	if err != nil {
		return err
	}

	for i, hit := range sr.Hits.Hits {
		var v map[string]any
		err = json.Unmarshal(hit, &v)
		if err != nil {
			return fmt.Errorf("failed to unmarshal hit %d: %w", i, err)
		}

		sID := v["_id"].(string)
		e := entry{v}

		cont, err := fn(sID, e)
		if !cont || err != nil {
			return err
		}
	}
	return nil
}

func (s *store) ensureEs() error {
	if !s.needES() {
		return nil
	}
	defer s.lock.Unlock()

	outputCfg := getOutputConfig()
	if outputCfg == nil {
		return errors.New("no output configuration")
	}

	es, err := eslegclient.NewConnectedClient(outputCfg, s.beatInfo.Beat)
	if err != nil {
		return fmt.Errorf("failed to create elasticsearch client: %w", err)
	}
	s.es = es
	return nil
}

func (s *store) hasES() bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.es != nil
}

func (s *store) needES() bool {
	if s.hasES() {
		return false
	}
	s.lock.Lock()
	if s.es != nil {
		s.lock.Unlock()
		return false
	}
	return true
}
