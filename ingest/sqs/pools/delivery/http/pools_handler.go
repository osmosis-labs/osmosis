package http

import (
	"net/http"
	"strconv"

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

const resourcePrefix = "/pools"

func formatPoolsResource(resource string) string {
	return resourcePrefix + resource
}

// NewPoolsHandler will initialize the pools/ resources endpoint
func NewPoolsHandler(e *echo.Echo, us mvc.PoolsUsecase) {
	handler := &PoolsHandler{
		PUsecase: us,
	}

	e.GET(formatPoolsResource("/all"), handler.GetAllPools)
	e.GET(formatPoolsResource("/:id"), handler.GetPool)
	e.GET(formatPoolsResource("/ticks/:id"), handler.GetConcentratedPoolTicks)
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

// GetPool will fetch a pool by its id
func (a *PoolsHandler) GetPool(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	poolID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
	}

	pools, err := a.PUsecase.GetPool(ctx, poolID)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, pools)
}

func (a *PoolsHandler) GetConcentratedPoolTicks(c echo.Context) error {
	ctx := c.Request().Context()

	idStr := c.Param("id")
	poolID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
	}

	pools, err := a.PUsecase.GetTickModelMap(ctx, []uint64{poolID})
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	tickModel, ok := pools[poolID]
	if !ok {
		return c.JSON(http.StatusNotFound, ResponseError{Message: "tick model not found for given pool"})
	}

	return c.JSON(http.StatusOK, tickModel)
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
