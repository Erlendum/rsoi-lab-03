package http

import (
	"context"
	"github.com/Erlendum/rsoi-lab-03/internal/library-system/config"
	"github.com/Erlendum/rsoi-lab-03/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

type libraryHandler interface {
	Register(echo *echo.Echo)
	GetLibraries(c echo.Context) error
	GetBooksByLibrary(c echo.Context) error
	GetBooksByUids(c echo.Context) error
	GetLibrariesByUids(c echo.Context) error
	UpdateBooksAvailableCount(c echo.Context) error
}

type server struct {
	echo           *echo.Echo
	cfg            *config.Server
	libraryHandler libraryHandler
}

func NewServer(cfg *config.Server, libraryHandler libraryHandler) *server {
	return &server{
		echo:           echo.New(),
		libraryHandler: libraryHandler,
		cfg:            cfg,
	}
}

func (s *server) Init() error {
	s.echo.Server.Addr = s.cfg.Address
	s.echo.HideBanner = true
	s.echo.HidePort = true

	s.echo.Use(
		middleware.CORSWithConfig(middleware.CORSConfig{
			AllowOrigins:                             []string{"*"},
			UnsafeWildcardOriginWithAllowCredentials: true,
			AllowCredentials:                         true,
		}),
	)

	s.echo.Validator = validation.MustRegisterCustomValidator(validator.New())

	s.echo.GET("/manage/health", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	s.libraryHandler.Register(s.echo)
	return nil
}

func (s *server) Run() error {
	log.Info().Msg("server has been started")
	return s.echo.StartServer(s.echo.Server)
}

func (s *server) Stop(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, s.cfg.ShutdownTimeout)
	defer cancel()
	if err := s.echo.Shutdown(ctx); err != nil {
		log.Err(err).Msg("could not stop server gracefully")
		return s.echo.Close()
	}
	return nil
}
