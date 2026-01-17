package runtime

import (
	"context"
	"errors"
	"net/http"
	"sync"
)

type runnableGroup interface {
	Start([]string) error
	Shutdown() error
}

type workerGroup struct {
	ctx     context.Context
	workers []*http.Server
	handler *dynamicHandler

	wg   sync.WaitGroup
	once sync.Once
}

func NewWorkerGroup(ctx context.Context, dh *dynamicHandler) runnableGroup {
	return &workerGroup{
		ctx:     ctx,
		handler: dh,
		workers: make([]*http.Server, 0),
	}
}

func (g *workerGroup) Start(listeners []string) error {
	// Create a http server for every listener port
	for _, addr := range listeners {
		g.workers = append(g.workers, &http.Server{Handler: g.handler, Addr: addr})
	}
	errCh := make(chan error, len(g.workers))
	// Start all the servers
	g.once.Do(func() {
		g.wg.Add(len(g.workers))
		for _, worker := range g.workers {
			go func() {
				defer g.wg.Done()
				// ListenAndServe will block until signalled by the Stop method
				if err := worker.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					errCh <- err
				}
			}()
		}
	})
	g.wg.Wait()
	close(errCh)

	var errList []error
	for err := range errCh {
		errList = append(errList, err)
	}

	if len(errList) != 0 {
		return errors.Join(errList...)
	}
	return nil
}

func (g *workerGroup) Shutdown() error {
	var errList []error
	for _, worker := range g.workers {
		if err := worker.Shutdown(g.ctx); err != nil {
			errList = append(errList, err)
		}
	}
	if len(errList) != 0 {
		return errors.Join(errList...)
	}
	return nil
}
