package rating

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strconv"
)

//go:generate mockgen -source=handler.go -destination=handler_mocks.go -self_package=github.com/Erlendum/rsoi-lab-02/internal/rating-system/rating -package=rating

type storage interface {
	CreateRatingRecord(ctx context.Context, record *ratingRecord) (int, error)
	UpdateRatingRecord(ctx context.Context, userName string, record *ratingRecord) error
	GetRatingRecord(ctx context.Context, username string) (ratingRecord, error)
}

type handler struct {
	storage storage
}

func NewHandler(storage storage) *handler {
	return &handler{storage: storage}
}

func (h *handler) Register(echo *echo.Echo) {
	api := echo.Group("/api/v1")

	api.GET("/rating/:username", h.GetRatingRecord)
	api.POST("/rating", h.CreateRatingRecord)
	api.PUT("/rating/:username", h.UpdateRatingRecord)
}

func (h *handler) GetRatingRecord(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "username is wrong"})
	}

	record, err := h.storage.GetRatingRecord(c.Request().Context(), username)
	if err != nil {
		log.Err(err).Msg("failed to get rating record")
		if errors.Is(err, errRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "record not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "storage error"})
	}

	type response struct {
		Stars int `json:"stars"`
	}

	return c.JSON(http.StatusOK, response{Stars: *record.Stars})
}

func (h *handler) CreateRatingRecord(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		log.Err(err).Msg("failed to read request body")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "failed to read request body"})
	}

	type request struct {
		UserName string `json:"userName" validate:"required"`
	}
	req := request{}

	err = json.Unmarshal(body, &req)
	if err != nil {
		log.Err(err).Msg("failed to unmarshal request body")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "failed to unmarshal request body"})
	}

	if err = c.Validate(req); err != nil {
		log.Err(err).Msg("failed to validate request body")
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "failed to validate request body"})
	}

	id, err := h.storage.CreateRatingRecord(c.Request().Context(), &ratingRecord{UserName: &req.UserName})
	if err != nil {
		log.Err(err).Msg("failed to create rating record")
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to create rating record"})
	}

	return c.JSON(http.StatusOK, echo.Map{"id": id})
}

func (h *handler) UpdateRatingRecord(c echo.Context) error {
	username := c.Param("username")
	if username == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "username is wrong"})
	}
	starsDiffParam := c.QueryParam("starsDiff")
	starsDiff, err := strconv.Atoi(starsDiffParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": "starsDiff is wrong"})
	}

	record, err := h.storage.GetRatingRecord(c.Request().Context(), username)
	if err != nil {
		log.Err(err).Msg("failed to get rating record")
		if errors.Is(err, errRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{"message": "record not found"})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "storage error"})
	}

	newStars := *record.Stars + starsDiff

	err = h.storage.UpdateRatingRecord(c.Request().Context(), username, &ratingRecord{Stars: &newStars})
	if err != nil {
		log.Err(err).Msg("failed to create rating record")
		return c.JSON(http.StatusInternalServerError, echo.Map{"message": "failed to create rating record"})
	}

	return c.NoContent(http.StatusOK)
}
