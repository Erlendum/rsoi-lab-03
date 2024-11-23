package library_system

import (
	"bytes"
	"errors"
	"github.com/Erlendum/rsoi-lab-03/internal/gateway/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type httpClientStub struct {
	err        error
	statusCode int
}

func (h *httpClientStub) Do(req *http.Request) (*http.Response, error) {
	return &http.Response{StatusCode: h.statusCode, Body: io.NopCloser(bytes.NewBufferString("test"))}, h.err
}

func Test_SendEvent(t *testing.T) {
	var tests = []struct {
		TestName string
		Data     struct {
			wantErr          error
			wantStatusCode   int
			expectedHTTPCode int
		}
	}{
		{
			TestName: "500 http-code",
			Data: struct {
				wantErr          error
				wantStatusCode   int
				expectedHTTPCode int
			}{wantErr: errors.New(""), expectedHTTPCode: http.StatusInternalServerError},
		},
		{
			TestName: "200 http-code",
			Data: struct {
				wantErr          error
				wantStatusCode   int
				expectedHTTPCode int
			}{wantStatusCode: http.StatusOK, wantErr: nil, expectedHTTPCode: http.StatusOK},
		},
	}

	e := echo.New()
	for _, tt := range tests {
		t.Run(tt.TestName, func(t *testing.T) {
			httpStub := httpClientStub{err: tt.Data.wantErr, statusCode: tt.Data.wantStatusCode}
			h := handler{httpClient: &httpStub, config: &config.Config{}}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rw := httptest.NewRecorder()
			c := e.NewContext(req, rw)

			err := h.GetLibraries(c)

			require.NoError(t, err)

			require.Equal(t, tt.Data.expectedHTTPCode, rw.Code)
		})
	}
}
