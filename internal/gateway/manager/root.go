package manager

import (
	"context"
	"github.com/Erlendum/rsoi-lab-03/internal/gateway/config"
	"github.com/Erlendum/rsoi-lab-03/internal/gateway/http"
	library_system "github.com/Erlendum/rsoi-lab-03/internal/gateway/library-system"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
)

type server interface {
	Init() error
	Run() error
	Stop(ctx context.Context) error
}

type root struct {
	errorChan chan error
	server    server
	cfg       *config.Config
}

func NewRoot() *root {
	return &root{}
}

func (r *root) Register(ctx context.Context) error {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Logger = log.With().Caller().Logger()

	var err error
	r.cfg, err = config.New()
	if err != nil {
		log.Error().Err(err).Msg("config load error")
		return err
	}

	librarySystemHandler := library_system.NewHandler(r.cfg)

	r.server = http.NewServer(&r.cfg.Server, librarySystemHandler)

	err = r.server.Init()
	if err != nil {
		log.Error().Err(err).Msg("server init error")
		return err
	}

	return nil
}

func (r *root) Resolve(ctx context.Context, shutdown chan os.Signal) os.Signal {
	go func() {
		log.Info().Msg("server started")
		r.errorChan <- r.server.Run()
	}()
	for {
		select {
		case err := <-r.errorChan:
			log.Err(err).Msg("error occurred")
		case sig := <-shutdown:
			return sig
		}
	}
}

func (r *root) Release(ctx context.Context, signal os.Signal) {
	log.Info().Msgf("shutdown started with signal : [%d]", signal)
	defer log.Info().Msg("shutdown completed")
	if err := r.server.Stop(ctx); err != nil {
		log.Err(err).Msg("could not stop server")
	}
}
