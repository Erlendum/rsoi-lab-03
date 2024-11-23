package reservation

import (
	"errors"
	"github.com/Erlendum/rsoi-lab-03/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type handlerTestFields struct {
	storage *Mockstorage
}

func createHandlerTestFields(ctrl *gomock.Controller) *handlerTestFields {
	return &handlerTestFields{
		storage: NewMockstorage(ctrl),
	}
}

func Test_UpdateReservationStatus(t *testing.T) {
	type fields struct {
		status           string
		username         string
		reservationUid   string
		expectedHTTPCode int
	}

	e := echo.New()
	e.Validator = validation.MustRegisterCustomValidator(validator.New())

	tests := []struct {
		name    string
		fields  fields
		Prepare func(fields *handlerTestFields)
	}{
		{
			name: "http-code 400: wrong status",
			fields: fields{
				expectedHTTPCode: http.StatusBadRequest,
				username:         "test",
				status:           "",
				reservationUid:   "test",
			},

			Prepare: func(fields *handlerTestFields) {
			},
		},
		{
			name: "http-code 400: wrong username",
			fields: fields{
				expectedHTTPCode: http.StatusBadRequest,
				username:         "",
				status:           "test",
				reservationUid:   "test",
			},

			Prepare: func(fields *handlerTestFields) {
			},
		},
		{
			name: "http-code 400: wrong reservationUid",
			fields: fields{
				expectedHTTPCode: http.StatusBadRequest,
				username:         "",
				status:           "test",
				reservationUid:   "",
			},

			Prepare: func(fields *handlerTestFields) {
			},
		},
		{
			name: "http-code 500: storage error",
			fields: fields{
				expectedHTTPCode: http.StatusInternalServerError,
				username:         "test",
				status:           "test",
				reservationUid:   "test",
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().UpdateReservationStatus(gomock.Any(), "test", "test", "test").Return(errors.New(""))
			},
		},
		{
			name: "http-code 404: not found error",
			fields: fields{
				expectedHTTPCode: http.StatusNotFound,
				username:         "test",
				status:           "test",
				reservationUid:   "test",
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().UpdateReservationStatus(gomock.Any(), "test", "test", "test").Return(errNotFound)
			},
		},
		{
			name: "http-code 200: success",
			fields: fields{
				expectedHTTPCode: http.StatusOK,
				username:         "test",
				status:           "test",
				reservationUid:   "test",
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().UpdateReservationStatus(gomock.Any(), "test", "test", "test").Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			testFields := createHandlerTestFields(ctrl)
			tt.Prepare(testFields)

			h := &handler{storage: testFields.storage}

			req := httptest.NewRequest(http.MethodPut, "/test?status="+tt.fields.status, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("uid")
			c.SetParamValues(tt.fields.reservationUid)
			c.Request().Header.Set("X-User-Name", tt.fields.username)

			err := h.UpdateReservationStatus(c)

			require.NoError(t, err)
			require.Equal(t, tt.fields.expectedHTTPCode, rec.Code)
		})
	}
}
