package http

import (
	"net/http"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// PoolsHandler  represent the httphandler for pools
type PoolsHandler struct {
	PUsecase mvc.PoolsUsecase
}

// NewPoolsHandler will initialize the pools/ resources endpoint
func NewPoolsHandler(e *echo.Echo, us mvc.PoolsUsecase) {
	handler := &PoolsHandler{
		PUsecase: us,
	}
	e.GET("/all-pools", handler.GetAllPools)
}

// GetAllPools will fetch all supported pool types by the Osmosis
// chain
func (a *PoolsHandler) GetAllPools(c echo.Context) error {
	ctx := c.Request().Context()

	pools, err := a.PUsecase.GetAllPools(ctx)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, pools)
}

func getStatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	logrus.Error(err)
	switch err {
	case domain.ErrInternalServerError:
		return http.StatusInternalServerError
	case domain.ErrNotFound:
		return http.StatusNotFound
	case domain.ErrConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
