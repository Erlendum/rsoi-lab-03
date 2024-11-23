package rating

import (
	"errors"
	"github.com/Erlendum/rsoi-lab-03/pkg/validation"
	"github.com/go-playground/validator/v10"
	"github.com/golang/mock/gomock"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"io"
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

func getPointerOnInt(i int) *int {
	return &i
}

func Test_GetRatingRecord(t *testing.T) {
	type fields struct {
		username             string
		expectedHTTPCode     int
		expectedResponseBody string
	}

	e := echo.New()
	e.Validator = validation.MustRegisterCustomValidator(validator.New())

	tests := []struct {
		name    string
		fields  fields
		Prepare func(fields *handlerTestFields)
	}{
		{
			name: "http-code 400: wrong username",
			fields: fields{
				expectedHTTPCode:     http.StatusBadRequest,
				username:             "",
				expectedResponseBody: ``,
			},

			Prepare: func(fields *handlerTestFields) {
			},
		},
		{
			name: "http-code 404: record not found",
			fields: fields{
				expectedHTTPCode:     http.StatusNotFound,
				username:             "test",
				expectedResponseBody: ``,
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().GetRatingRecord(gomock.Any(), "test").Return(ratingRecord{}, errRecordNotFound)
			},
		},
		{
			name: "http-code 500: storage error",
			fields: fields{
				expectedHTTPCode:     http.StatusInternalServerError,
				username:             "test",
				expectedResponseBody: ``,
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().GetRatingRecord(gomock.Any(), "test").Return(ratingRecord{}, errors.New(""))
			},
		},
		{
			name: "http-code 200: success",
			fields: fields{
				expectedHTTPCode: http.StatusOK,
				username:         "test",
				expectedResponseBody: `{"stars":100}
`,
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().GetRatingRecord(gomock.Any(), "test").Return(ratingRecord{Stars: getPointerOnInt(100)}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			testFields := createHandlerTestFields(ctrl)
			tt.Prepare(testFields)

			h := &handler{storage: testFields.storage}

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("username")
			c.SetParamValues(tt.fields.username)

			err := h.GetRatingRecord(c)

			require.NoError(t, err)
			require.Equal(t, tt.fields.expectedHTTPCode, rec.Code)
			if tt.fields.expectedHTTPCode == http.StatusOK {
				body, err := io.ReadAll(rec.Result().Body)
				require.NoError(t, err)
				require.Equal(t, tt.fields.expectedResponseBody, string(body))
			}
		})
	}
}
