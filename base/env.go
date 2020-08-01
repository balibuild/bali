package base

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"sync"
)

// Derivator expand env engine
type Derivator struct {
	envBlock map[string]string
	mu       sync.RWMutex
}

// NewDerivator create env derivative
func NewDerivator() *Derivator {
	return &Derivator{
		envBlock: make(map[string]string),
	}
}

// AddBashCompatible $0~$9
func (de *Derivator) AddBashCompatible() {
	de.mu.Lock()
	defer de.mu.Unlock()
	for i := 0; i < len(os.Args); i++ {
		de.envBlock[strconv.Itoa(i)] = os.Args[i]
	}
	de.envBlock["$"] = strconv.Itoa(os.Getpid())
}

// Append append to env
func (de *Derivator) Append(k, v string) error {
	if k == "" || v == "" {
		return errors.New("empty env k/v input")
	}
	de.mu.Lock()
	defer de.mu.Unlock()
	de.envBlock[k] = v
	return nil
}

// Environ create new environ block
func (de *Derivator) Environ() []string {
	de.mu.RLock()
	defer de.mu.RUnlock()
	oe := os.Environ()
	ev := make([]string, 0, len(oe)+len(de.envBlock))
	for _, e := range oe {
		kv := strings.Split(e, "=")
		if len(kv) > 0 {
			k := kv[0]
			if _, ok := de.envBlock[k]; ok {
				continue
			}
		}
		ev = append(ev, e)
	}
	for k, v := range de.envBlock {
		ev = append(ev, StrCat(k, "=", v))
	}
	return ev
}

// EraseEnv k
func (de *Derivator) EraseEnv(k string) {
	de.mu.Lock()
	defer de.mu.Unlock()
	delete(de.envBlock, k)
}

// GetEnv env
func (de *Derivator) GetEnv(k string) string {
	de.mu.RLock()
	defer de.mu.RUnlock()
	if v, ok := de.envBlock[k]; ok {
		return v
	}
	return os.Getenv(k)
}

// ExpandEnv env
func (de *Derivator) ExpandEnv(s string) string {
	return os.Expand(s, de.GetEnv)
}
