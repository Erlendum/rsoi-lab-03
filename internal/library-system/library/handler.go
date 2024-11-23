package library

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

//go:generate mockgen -source=handler.go -destination=handler_mocks.go -self_package=github.com/Erlendum/rsoi-lab-02/internal/library-system/library -package=library

type storage interface {
	GetLibraries(ctx context.Context, city string, offset, limit int) ([]library, error)
	GetBooksByLibrary(ctx context.Context, libraryUid string, offset, limit int, showAll bool) ([]book, error)
	GetBooksAvailableCount(ctx context.Context, libraryUid, bookUid string) (int, error)
	GetBooksByUids(ctx context.Context, uids []string) ([]book, error)
	GetLibrariesByUids(ctx context.Context, uids []string) ([]library, error)
	UpdateBooksAvailableCount(ctx context.Context, libraryUid, bookUid string, count int) error
}

type handler struct {
	storage storage
}

func NewHandler(storage storage) *handler {
	return &handler{storage: storage}
}

func (h *handler) Register(echo *echo.Echo) {
	api := echo.Group("/api/v1")

	api.GET("/libraries", h.GetLibraries)
	api.GET("/libraries/:uid/books", h.GetBooksByLibrary)
	api.GET("/books/", h.GetBooksByUids)
	api.GET("/libraries/by-uids", h.GetLibrariesByUids)
	api.PUT("/libraries/:libraryuid/books/:bookuid", h.UpdateBooksAvailableCount)
}

func (h *handler) GetLibraries(c echo.Context) error {
	city := c.QueryParam("city")
	if city == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "city is wrong",
		})
	}

	pageParam := c.QueryParam("page")
	page, err := strconv.Atoi(pageParam)
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "page is wrong",
		})
	}

	sizeParam := c.QueryParam("size")
	size, err := strconv.Atoi(sizeParam)
	if err != nil || size <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "size is wrong",
		})
	}

	libraries, err := h.storage.GetLibraries(c.Request().Context(), city, page*size-size, size)

	if err != nil {
		log.Err(err).Msg("failed to get libraries")
		if errors.Is(err, errLibraryNotFound) {
			return c.NoContent(http.StatusNoContent)
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to get libraries",
		})
	}

	type item struct {
		LibraryUid string `json:"libraryUid"`
		Name       string `json:"name"`
		Address    string `json:"address"`
		City       string `json:"city"`
	}
	type response struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"pageSize"`
		TotalElements int    `json:"totalElements"`
		Items         []item `json:"items"`
	}

	items := make([]item, 0, len(libraries))
	for _, v := range libraries {
		items = append(items, item{
			LibraryUid: v.LibraryUid,
			Name:       v.Name,
			Address:    v.Address,
			City:       v.City,
		})
	}
	res := response{
		Page:          page,
		PageSize:      size,
		TotalElements: len(libraries),
		Items:         items,
	}

	return c.JSON(http.StatusOK, res)

}

func (h *handler) GetBooksByLibrary(c echo.Context) error {
	libraryUid := c.Param("uid")
	if libraryUid == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "uid is wrong",
		})
	}

	pageParam := c.QueryParam("page")
	page, err := strconv.Atoi(pageParam)
	if err != nil || page <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "page is wrong",
		})
	}

	sizeParam := c.QueryParam("size")
	size, err := strconv.Atoi(sizeParam)
	if err != nil || size <= 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "size is wrong",
		})
	}

	showAllParam := c.QueryParam("showAll")
	showAll, err := strconv.ParseBool(showAllParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "showAll is wrong",
		})
	}

	books, err := h.storage.GetBooksByLibrary(c.Request().Context(), libraryUid, page*size-size, size, showAll)

	if err != nil {
		log.Err(err).Msg("failed to get books")
		if errors.Is(err, errBookNotFound) {
			return c.NoContent(http.StatusNoContent)
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to get books",
		})
	}

	type item struct {
		BookUid        string `json:"bookUid"`
		Name           string `json:"name"`
		Author         string `json:"author"`
		Genre          string `json:"genre"`
		Condition      string `json:"condition"`
		AvailableCount int    `json:"availableCount"`
	}
	type response struct {
		Page          int    `json:"page"`
		PageSize      int    `json:"pageSize"`
		TotalElements int    `json:"totalElements"`
		Items         []item `json:"items"`
	}

	items := make([]item, 0, len(books))
	for _, v := range books {
		items = append(items, item{
			BookUid:        v.BookUid,
			Name:           v.Name,
			Author:         v.Author,
			Genre:          v.Genre,
			Condition:      v.Condition,
			AvailableCount: v.AvailableCount,
		})
	}
	res := response{
		Page:          page,
		PageSize:      size,
		TotalElements: len(books),
		Items:         items,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *handler) GetBooksByUids(c echo.Context) error {
	uids := c.QueryParams()["bookUids"]
	if len(uids) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "uids are wrong",
		})
	}

	books, err := h.storage.GetBooksByUids(c.Request().Context(), uids)

	if err != nil {
		log.Err(err).Msg("failed to get books")
		if errors.Is(err, errBookNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to get books",
		})
	}

	type item struct {
		BookUid   string `json:"bookUid"`
		Name      string `json:"name"`
		Author    string `json:"author"`
		Genre     string `json:"genre"`
		Condition string `json:"condition"`
	}
	type response struct {
		Data []item `json:"data"`
	}

	items := make([]item, 0, len(books))
	for _, v := range books {
		items = append(items, item{
			BookUid:   v.BookUid,
			Name:      v.Name,
			Author:    v.Author,
			Genre:     v.Genre,
			Condition: v.Condition,
		})
	}

	res := response{
		Data: items,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *handler) GetLibrariesByUids(c echo.Context) error {
	uids := c.QueryParams()["libraryUids"]
	if len(uids) == 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "uids are wrong",
		})
	}

	libraries, err := h.storage.GetLibrariesByUids(c.Request().Context(), uids)

	if err != nil {
		log.Err(err).Msg("failed to get libraries")
		if errors.Is(err, errLibraryNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to get libraries",
		})
	}

	type item struct {
		LibraryUid string `json:"libraryUid"`
		Name       string `json:"name"`
		Address    string `json:"address"`
		City       string `json:"city"`
	}
	type response struct {
		Data []item `json:"data"`
	}

	items := make([]item, 0, len(libraries))
	for _, v := range libraries {
		items = append(items, item{
			LibraryUid: v.LibraryUid,
			Name:       v.Name,
			Address:    v.Address,
			City:       v.City,
		})
	}

	res := response{
		Data: items,
	}

	return c.JSON(http.StatusOK, res)
}

func (h *handler) UpdateBooksAvailableCount(c echo.Context) error {
	libraryUid := c.Param("libraryuid")
	bookUid := c.Param("bookuid")
	if libraryUid == "" || bookUid == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "uid is wrong",
		})
	}

	countDiffParam := c.QueryParam("countDiff")
	countDiff, err := strconv.Atoi(countDiffParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "count is wrong",
		})
	}

	actualCount, err := h.storage.GetBooksAvailableCount(c.Request().Context(), libraryUid, bookUid)
	if err != nil {
		log.Err(err).Msg("failed to get available count")
		if errors.Is(err, errRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"message": "record not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to get available count",
		})
	}

	if actualCount+countDiff < 0 {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"message": "count is wrong",
		})
	}

	err = h.storage.UpdateBooksAvailableCount(c.Request().Context(), libraryUid, bookUid, actualCount+countDiff)
	if err != nil {
		log.Err(err).Msg("failed to update books available count")
		if errors.Is(err, errRecordNotFound) {
			return c.JSON(http.StatusNotFound, echo.Map{
				"message": "record not found",
			})
		}
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to update books available count",
		})
	}

	return c.NoContent(http.StatusOK)
}
