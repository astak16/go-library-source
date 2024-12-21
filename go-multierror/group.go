package multierror

import (
	"sync"
)

type Group struct {
	mutex sync.Mutex
	err   *Error
	wg    sync.WaitGroup
}

func (g *Group) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		if err := f(); err != nil {
			g.mutex.Lock()
			g.err = Append(g.err, err)
			g.mutex.Unlock()
		}
	}()
}

func (g *Group) Wait() *Error {
	g.wg.Wait()
	g.mutex.Lock()
	defer g.mutex.Unlock()
	return g.err
}
