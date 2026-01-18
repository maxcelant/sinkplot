package runtime

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

type runnableGroup interface {
	Start([]int) error
	Shutdown(context.Context) error
}

type workerGroup struct {
	workers []*http.Server
	handler *dynamicHandler

	wg   sync.WaitGroup
	once sync.Once
}

func NewWorkerGroup(dh *dynamicHandler) runnableGroup {
	return &workerGroup{
		handler: dh,
		workers: make([]*http.Server, 0),
	}
}

func (g *workerGroup) Start(listeners []int) error {
	// Create a http server for every listener port
	for _, addr := range listeners {
		log.Info().Msgf("starting worker server on :%d", addr)
		g.workers = append(g.workers, &http.Server{Handler: g.handler, Addr: fmt.Sprintf(":%d", addr)})
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

func (g *workerGroup) Shutdown(shutdownCtx context.Context) error {
	var errList []error
	for _, worker := range g.workers {
		if err := worker.Shutdown(shutdownCtx); err != nil {
			errList = append(errList, err)
		}
	}
	if len(errList) != 0 {
		return errors.Join(errList...)
	}
	return nil
}
