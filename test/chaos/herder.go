package chaos

import (
	"context"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
	pkgupgrade "knative.dev/pkg/test/upgrade"
)

type Herder interface {
	LookAfter(duck Duck)
	Herd()
	channelRetriver
}

// NewHerder creates a field on which Duck's could Quak at K8s resources.
func NewHerder(upgradeCtx pkgupgrade.Context, cfg Config) Herder {
	return &doroty{
		stateCtx: &stateCtx{
			ctx: context.Background(),
			cfg: cfg,
			tb:  upgradeCtx.T,
			log: upgradeCtx.Log,
		},
		finishCh: make(chan *sync.WaitGroup),
		readyWg:  &sync.WaitGroup{},
	}
}

type channelRetriver interface {
	finished() <-chan *sync.WaitGroup
	ready() *sync.WaitGroup
}

type doroty struct {
	*stateCtx
	finishCh chan *sync.WaitGroup
	readyWg  *sync.WaitGroup
	ducks    []Duck
}

// LookAfter a duck so it can Quak at some resources in the future.
func (d *doroty) LookAfter(duck Duck) {
	d.ducks = append(d.ducks,
		duck.takenCareOf(d.stateCtx, d),
	)
}

func (d *doroty) Herd() {
	log := d.log.With("name", "Doroty")
	log.Infof("Herding %d ducks...", len(d.ducks))
	d.readyWg.Add(len(d.ducks))
	for _, duck := range d.ducks {
		go duck.Quak()
	}
	waitOrFail(d.tb, d.readyWg, "start quaking", time.Minute)
	d.awaitAndSleep(log)
	finishedWg := &sync.WaitGroup{}
	finishedWg.Add(len(d.ducks))
	for i := 1; i <= len(d.ducks); i++ {
		d.finishCh <- finishedWg
	}
	waitOrFail(d.tb, finishedWg, "finish quaking", time.Minute)
	close(d.finishCh)
}

func (d *doroty) awaitAndSleep(log *zap.SugaredLogger) {
	t := d.cfg.Timeout
	log.Infof("Waiting for %v until ducks do their thing...", t)
	for t > 0 {
		w := d.cfg.Wait
		if w > t {
			w = t
		}
		time.Sleep(w)
		t = t - w
		if t > 0 {
			log.Infof("Is it yet? No, there's still %v left. Going back to sleep...", t)
		}
	}
	log.Info("That's enough. Stop quaking!")
}

func waitOrFail(tb testing.TB, wg *sync.WaitGroup, thing string, timeout time.Duration) {
	if !waitTimeout(wg, timeout) {
		tb.Fatalf("Wait timed out (%v) for: %s", timeout, thing)
	}
}

// waitTimeout waits for the sync.WaitGroup for the specified max timeout.
// Returns true if waiting was successful.
// Ref.: https://stackoverflow.com/a/32843750/844449
func waitTimeout(wg *sync.WaitGroup, timeout time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		// completed normally
		return true
	case <-time.After(timeout):
		// timed out
		return false
	}
}

func (d *doroty) finished() <-chan *sync.WaitGroup {
	return d.finishCh
}

func (d *doroty) ready() *sync.WaitGroup {
	return d.readyWg
}
