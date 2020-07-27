package errcol

import "sync"

type Collector interface {
	Err(err error)
	ForEach(func(error))
}

type DefaultCollector struct {
	mu   sync.RWMutex
	errs []error
}

func Default() *DefaultCollector {
	return &DefaultCollector{}
}

func (c *DefaultCollector) Err(err error) {
	if err == nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.errs = append(c.errs, err)
}

func (c *DefaultCollector) ForEach(f func(error)) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	for _, err := range c.errs {
		f(err)
	}
}
