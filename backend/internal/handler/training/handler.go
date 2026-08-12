package traininghandler

import (
	"errors"
	"net/http"
	"strconv"

	"atcoder_shojin/backend/internal/domain/training"
	trainingusecase "atcoder_shojin/backend/internal/usecase/training"
	"github.com/labstack/echo/v4"
)

type Handler struct{ usecase *trainingusecase.Usecase }

func New(usecase *trainingusecase.Usecase) *Handler { return &Handler{usecase: usecase} }

func (handler *Handler) Start(context echo.Context) error {
	response, err := handler.usecase.Start(context.Request().Context())
	if err != nil {
		return writeError(context, err)
	}
	return context.JSON(http.StatusCreated, response)
}
func (handler *Handler) Active(context echo.Context) error {
	response, err := handler.usecase.Active(context.Request().Context())
	if err != nil {
		return writeError(context, err)
	}
	if response.Session == nil {
		return context.JSON(http.StatusNotFound, map[string]string{"error": "No active training session."})
	}
	return context.JSON(http.StatusOK, response)
}
func (handler *Handler) Get(context echo.Context) error {
	response, err := handler.usecase.Get(context.Request().Context(), context.Param("id"))
	if err != nil {
		return writeError(context, err)
	}
	return context.JSON(http.StatusOK, response)
}
func (handler *Handler) Sync(context echo.Context) error {
	response, err := handler.usecase.Sync(context.Request().Context(), context.Param("id"))
	if err != nil {
		return writeError(context, err)
	}
	return context.JSON(http.StatusOK, response)
}
func (handler *Handler) Abort(context echo.Context) error {
	response, err := handler.usecase.Abort(context.Request().Context(), context.Param("id"))
	if err != nil {
		return writeError(context, err)
	}
	return context.JSON(http.StatusOK, response)
}
func (handler *Handler) History(context echo.Context) error {
	page := 1
	if raw := context.QueryParam("page"); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return context.JSON(http.StatusBadRequest, map[string]string{"error": "page must be a positive integer"})
		}
		page = value
	}
	response, err := handler.usecase.History(context.Request().Context(), page)
	if err != nil {
		return writeError(context, err)
	}
	return context.JSON(http.StatusOK, response)
}

func writeError(context echo.Context, err error) error {
	status := http.StatusInternalServerError
	message := "Internal server error."
	switch {
	case errors.Is(err, training.ErrProblemSetUnavailable):
		status = http.StatusUnprocessableEntity
		message = "Could not generate a valid problem set."
	case errors.Is(err, training.ErrActiveSessionExists):
		status = http.StatusConflict
		message = "An active training session already exists."
	case errors.Is(err, training.ErrAbortCooldown):
		status = http.StatusConflict
		message = "A new training cannot start before the aborted session deadline."
	case errors.Is(err, training.ErrExternalDataStale):
		status = http.StatusServiceUnavailable
		message = "Required AtCoder data could not be refreshed."
	case errors.Is(err, training.ErrSessionNotFound):
		status = http.StatusNotFound
		message = "Training session not found."
	case errors.Is(err, training.ErrSessionNotActive):
		status = http.StatusConflict
		message = "Training session is not active."
	}
	return context.JSON(status, map[string]string{"error": message})
}
