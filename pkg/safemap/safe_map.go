package safemap

import "sync"

type SafeMap struct {
	m   map[interface{}]interface{}
	mux *sync.Mutex
}

func New() SafeMap {
	return SafeMap{
		m:   make(map[interface{}]interface{}),
		mux: new(sync.Mutex),
	}
}

func (m SafeMap) Set(key, value interface{}) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.m[key] = value
}

func (m SafeMap) Get(key interface{}) (interface{}, bool) {
	m.mux.Lock()
	defer m.mux.Unlock()
	value, ok := m.m[key]
	return value, ok
}

func (m SafeMap) HasKey(key interface{}) bool {
	_, hasKey := m.Get(key)
	return hasKey
}

func (m SafeMap) Remove(key interface{}) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.m, key)
}
