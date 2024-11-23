package library

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

func Test_UpdateBooksAvailableCount(t *testing.T) {
	type fields struct {
		libraryUid       string
		bookUid          string
		countDiff        string
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
			name: "http-code 400: wrong libraryuid",
			fields: fields{
				expectedHTTPCode: http.StatusBadRequest,
				libraryUid:       "",
				bookUid:          "test",
				countDiff:        "1",
			},

			Prepare: func(fields *handlerTestFields) {
			},
		},
		{
			name: "http-code 400: wrong bookuid",
			fields: fields{
				expectedHTTPCode: http.StatusBadRequest,
				libraryUid:       "test",
				bookUid:          "",
				countDiff:        "1",
			},

			Prepare: func(fields *handlerTestFields) {
			},
		},
		{
			name: "http-code 400: wrong countDiff - not integer",
			fields: fields{
				expectedHTTPCode: http.StatusBadRequest,
				libraryUid:       "test",
				bookUid:          "test",
				countDiff:        "test",
			},

			Prepare: func(fields *handlerTestFields) {
			},
		},
		{
			name: "http-code 400: wrong countDiff - negative available count",
			fields: fields{
				expectedHTTPCode: http.StatusBadRequest,
				libraryUid:       "test",
				bookUid:          "test",
				countDiff:        "-2",
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().GetBooksAvailableCount(gomock.Any(), "test", "test").Return(1, nil)
			},
		},
		{
			name: "http-code 500: GetBooksAvailableCount error",
			fields: fields{
				expectedHTTPCode: http.StatusInternalServerError,
				libraryUid:       "test",
				bookUid:          "test",
				countDiff:        "-1",
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().GetBooksAvailableCount(gomock.Any(), "test", "test").Return(0, errors.New(""))
			},
		},
		{
			name: "http-code 500: UpdateBooksAvailableCount error",
			fields: fields{
				expectedHTTPCode: http.StatusInternalServerError,
				libraryUid:       "test",
				bookUid:          "test",
				countDiff:        "-1",
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().GetBooksAvailableCount(gomock.Any(), "test", "test").Return(1, nil)
				fields.storage.EXPECT().UpdateBooksAvailableCount(gomock.Any(), "test", "test", 0).Return(errors.New(""))
			},
		},
		{
			name: "http-code 200: success",
			fields: fields{
				expectedHTTPCode: http.StatusOK,
				libraryUid:       "test",
				bookUid:          "test",
				countDiff:        "-1",
			},

			Prepare: func(fields *handlerTestFields) {
				fields.storage.EXPECT().GetBooksAvailableCount(gomock.Any(), "test", "test").Return(1, nil)
				fields.storage.EXPECT().UpdateBooksAvailableCount(gomock.Any(), "test", "test", 0).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)

			testFields := createHandlerTestFields(ctrl)
			tt.Prepare(testFields)

			h := &handler{storage: testFields.storage}

			req := httptest.NewRequest(http.MethodPut, "/test?countDiff="+tt.fields.countDiff, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetParamNames("libraryuid", "bookuid")
			c.SetParamValues(tt.fields.libraryUid, tt.fields.bookUid)

			err := h.UpdateBooksAvailableCount(c)

			require.NoError(t, err)
			require.Equal(t, tt.fields.expectedHTTPCode, rec.Code)
		})
	}
}
