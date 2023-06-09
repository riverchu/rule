package web

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/riverchu/pkg/log"
	"github.com/riverchu/pkg/shutdown"
	"github.com/riverchu/rule"
	"github.com/riverchu/rule/driver"
)

const treeName = "default"

var f *rule.Forest

func Serve(timeout time.Duration, handler http.Handler) {
	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() { // service connections
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("listen: %s", err)
		}
	}()

	// wati shutdown signal
	<-shutdown.Cancel()
	log.Info("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	// catching ctx.Done(). timeout of 5 seconds.
	select {
	case <-ctx.Done():
		log.Error("timeout of %s.", timeout)
	}
	log.Info("Server exiting")

}

func DefaultBuilder(load func() []*rule.Rule) rule.TreeBuilder {
	return func() (string, *rule.Tree) {
		tree, err := rule.NewTree(&webDriver{PathParser: driver.SlashPathParser},
			treeName, `{}`, load()...)
		if err != nil {
			panic(fmt.Errorf("build new tree fail: %w", err))
		}
		return treeName, tree
	}
}

func InitForest(builders ...rule.TreeBuilder) { f = rule.NewForest(builders...) }

func RefreshForest() { f = f.Build() }

type webDriver struct {
	driver.PathParser

	driver.StdCalculator
	driver.DummyOperatorModem
}

func (webDriver) Name() string { return "default" }
