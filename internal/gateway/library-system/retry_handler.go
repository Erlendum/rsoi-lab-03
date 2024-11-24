package library_system

import (
	"github.com/RohanPoojary/gomq"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

const (
	defaultRetryTimeout = 10 * time.Second
)

type retryData struct {
	Time    time.Time
	Call    func(c echo.Context) error
	Context echo.Context
	Params  map[string]string
}

type retryHandler struct {
	broker  gomq.Broker
	Timeout time.Duration
}

func NewRetryHandler() *retryHandler {
	return &retryHandler{
		broker:  gomq.NewBroker(),
		Timeout: defaultRetryTimeout,
	}
}

func (h *retryHandler) Handle() {
	poller := h.broker.Subscribe(gomq.ExactMatcher("request.retry"))
	go func() {
		for {
			value, ok := poller.Poll()
			if !ok {
				return
			}

			data, ok := value.(retryData)
			if !ok {
				log.Error().Msg("invalid request.retry message type")
				continue
			}

			names := make([]string, 0, len(data.Params))
			values := make([]string, 0, len(data.Params))

			for k, v := range data.Params {
				names = append(names, k)
				values = append(values, v)
			}

			data.Context.SetParamNames(names...)
			data.Context.SetParamValues(values...)

			log.Info().Msgf("poller message from request.retry: %v", data)
			for time.Now().Sub(data.Time) <= h.Timeout {
			}

			err := data.Call(data.Context)
			if err != nil {
				log.Error().Err(err).Msg("failed to retry request")
				h.broker.Publish("request.retry", retryData{
					Time:    time.Now(),
					Call:    data.Call,
					Context: data.Context,
					Params:  data.Params,
				})
				continue
			}

			if data.Context.Response().Status >= http.StatusInternalServerError {
				log.Error().Err(err).Msg("failed to retry request")
				h.broker.Publish("request.retry", retryData{
					Time:    time.Now(),
					Call:    data.Call,
					Context: data.Context,
					Params:  data.Params,
				})
				continue
			}
		}
	}()
}
