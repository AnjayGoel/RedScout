package scanner

import (
	"sync"

	"redscout/models"
)

type keySample struct {
	key    string
	memory int64
	ttl    int64
	typ    string
}

type keyCommand struct {
	key     string
	command string
}

type opCounts map[keyCommand]int64

type store struct {
	mu      sync.RWMutex
	samples []keySample
	ops     opCounts
}

func newStore() *store {
	return &store{ops: opCounts{}}
}

func (st *store) resetSamples() {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.samples = nil
}

func (st *store) addSamples(samples []keySample) {
	st.mu.Lock()
	defer st.mu.Unlock()
	st.samples = append(st.samples, samples...)
}

func (st *store) addOps(ops opCounts) {
	st.mu.Lock()
	defer st.mu.Unlock()
	for kc, n := range ops {
		st.ops[kc] += n
	}
}

func (st *store) eachSample(fn func(keySample)) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	for _, s := range st.samples {
		fn(s)
	}
}

func (st *store) eachOp(fn func(kc keyCommand, count int64)) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	for kc, n := range st.ops {
		fn(kc, n)
	}
}

func (st *store) firstKeyWithPrefix(kp *models.KeyParser, prefix models.Key) (string, bool) {
	st.mu.RLock()
	defer st.mu.RUnlock()
	for _, s := range st.samples {
		if kp.IsA(kp.NewKey(s.key, false), prefix) {
			return s.key, true
		}
	}
	return "", false
}
