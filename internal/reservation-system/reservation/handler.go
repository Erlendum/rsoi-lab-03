package reservation

import (
	"context"
	"encoding/json"
	"errors"
	my_time "github.com/Erlendum/rsoi-lab-03/pkg/time"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"time"
)

var (
	rentedStatus = "RENTED"
)

//go:generate mockgen -source=handler.go -destination=handler_mocks.go -self_package=github.com/Erlendum/rsoi-lab-02/internal/reservation-system/reservation -package=reservation

type storage interface {
	CreateReservation(ctx context.Context, r *reservation) (int, error)
	UpdateReservationStatus(ctx context.Context, uid string, username string, status string) error
	GetReservation(ctx context.Context, uid string) (reservation, error)
	GetReservations(ctx context.Context, username string, status string) ([]reservation, error)
}

type handler struct {
	storage storage
}

func NewHandler(storage storage) *handler {
	return &handler{storage: storage}
}

func (h *handler) Register(echo *echo.Echo) {
	api := echo.Group("/api/v1")

	api.GET("/reservations/by-user/:username", h.GetReservations)
	api.GET("/reservations/:uid", h.GetReservationByUid)
	api.POST("/reservations/", h.CreateReservation)
	api.PUT("/reservations/:uid/status", h.UpdateReservationStatus)
}

func (h *handler) GetReservations(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "username is wrong",
		})
	}

	status := c.QueryParam("status")
	if status == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "status is wrong",
		})
	}

	r, err := h.storage.GetReservations(c.Request().Context(), username, status)
	if err != nil {
		log.Err(err).Msg("failed to get reservations")
		if errors.Is(err, errNotFound) {
			return c.JSON(http.StatusNoContent, echo.Map{
				"message": "reservations not found",
			})
		}
	}

	type response struct {
		ReservationUid string `json:"reservationUid"`
		Status         string `json:"status"`
		StartDate      string `json:"startDate"`
		TillDate       string `json:"tillDate"`
		BookUid        string `json:"bookUid"`
		LibraryUid     string `json:"libraryUid"`
	}

	res := make([]response, 0, len(r))
	for _, v := range r {
		res = append(res, response{
			ReservationUid: *v.ReservationUid,
			Status:         *v.Status,
			StartDate:      v.StartDate.String(),
			TillDate:       v.TillDate.String(),
			BookUid:        *v.BookUid,
			LibraryUid:     *v.LibraryUid,
		})
	}

	return c.JSON(http.StatusOK, res)
}

func (h *handler) GetReservationByUid(c echo.Context) error {
	uid := c.Param("uid")
	if uid == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "uid is wrong",
		})
	}

	r, err := h.storage.GetReservation(c.Request().Context(), uid)

	if err != nil {
		log.Err(err).Msg("failed to get reservation")
		if errors.Is(err, errNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"message": "reservation not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to get reservation",
		})
	}

	type response struct {
		ReservationUid string `json:"reservationUid"`
		Status         string `json:"status"`
		StartDate      string `json:"startDate"`
		TillDate       string `json:"tillDate"`
		BookUid        string `json:"bookUid"`
		LibraryUid     string `json:"libraryUid"`
	}

	return c.JSON(http.StatusOK, response{
		ReservationUid: *r.ReservationUid,
		Status:         *r.Status,
		StartDate:      r.StartDate.String(),
		TillDate:       r.TillDate.String(),
		BookUid:        *r.BookUid,
		LibraryUid:     *r.LibraryUid,
	})
}

func (h *handler) CreateReservation(c echo.Context) error {
	username := c.Request().Header.Get("X-User-Name")
	if username == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "username is wrong",
		})
	}

	type request struct {
		BookUid    string       `json:"bookUid" validate:"required"`
		LibraryUid string       `json:"libraryUid" validate:"required"`
		TillDate   my_time.Date `json:"tillDate" validate:"required"`
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Err(err).Msg("failed to read body")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to read body",
		})
	}
	req := &request{}

	if err = json.Unmarshal(body, &req); err != nil {
		log.Err(err).Msg("failed to unmarshal body")
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "failed to unmarshal body",
		})
	}

	now := my_time.Date(time.Now())
	reservationUid := uuid.New().String()
	_, err = h.storage.CreateReservation(c.Request().Context(), &reservation{
		BookUid:        &req.BookUid,
		ReservationUid: &reservationUid,
		LibraryUid:     &req.LibraryUid,
		TillDate:       &req.TillDate,
		StartDate:      &now,
		UserName:       &username,
		Status:         &rentedStatus,
	})

	if err != nil {
		log.Err(err).Msg("failed to create reservation")
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to create reservation",
		})
	}

	type response struct {
		ReservationUid string `json:"reservationUid"`
		Status         string `json:"status"`
		StartDate      string `json:"startDate"`
		TillDate       string `json:"tillDate"`
		BookUid        string `json:"bookUid"`
		LibraryUid     string `json:"libraryUid"`
	}

	return c.JSON(http.StatusOK, response{
		ReservationUid: reservationUid,
		Status:         rentedStatus,
		StartDate:      now.String(),
		TillDate:       req.TillDate.String(),
		BookUid:        req.BookUid,
		LibraryUid:     req.LibraryUid,
	})
}

func (h *handler) UpdateReservationStatus(c echo.Context) error {
	username := c.Request().Header.Get("X-User-Name")
	if username == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "username is wrong",
		})
	}

	status := c.QueryParam("status")
	if status == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "status is wrong",
		})
	}

	uid := c.Param("uid")
	if uid == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "uid is wrong",
		})
	}

	err := h.storage.UpdateReservationStatus(c.Request().Context(), uid, username, status)
	if err != nil {
		log.Err(err).Msg("failed to update reservation status")
		if errors.Is(err, errNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"message": "reservation not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to update reservation status",
		})
	}

	return c.NoContent(http.StatusOK)
}
