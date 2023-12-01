package http

import (
	"errors"
	"net/http"
	"regexp"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/domain/mvc"
	"github.com/osmosis-labs/osmosis/v21/ingest/sqs/log"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// RouterHandler  represent the httphandler for the router
type RouterHandler struct {
	RUsecase mvc.RouterUsecase
	logger   log.Logger
}

// Define a regular expression pattern to match sdk.Coin where the first part is the amount and second is the denom name
// Patterns tested:
// 500ibc/27394FB092D2ECCD56123C74F36E4C1F926001CEADA9CA97EA622B25F41E5EB2
// 100uion
var coinPattern = regexp.MustCompile(`([0-9]+)(([a-z]+)(\/([A-Z0-9]+))*)`)

// NewRouterHandler will initialize the pools/ resources endpoint
func NewRouterHandler(e *echo.Echo, us mvc.RouterUsecase, logger log.Logger) {
	handler := &RouterHandler{
		RUsecase: us,
		logger:   logger,
	}
	e.GET("/quote", handler.GetOptimalQuote)
	e.GET("/single-quote", handler.GetBestSingleRouteQuote)
	e.GET("/routes", handler.GetCandidateRoutes)
	e.POST("/store-state", handler.StoreRouterStateInFiles)
}

// GetOptimalQuote will determine the optimal quote for a given tokenIn and tokenOutDenom
// Return the optimal quote.
func (a *RouterHandler) GetOptimalQuote(c echo.Context) error {
	ctx := c.Request().Context()

	tokenOutDenom, tokenIn, err := getValidRoutingParameters(c)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	quote, err := a.RUsecase.GetOptimalQuote(ctx, tokenIn, tokenOutDenom)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	quote.PrepareResult()

	err = c.JSON(http.StatusOK, quote)
	if err != nil {
		return err
	}

	return nil
}

// GetBestSingleRouteQuote returns the best single route quote to be done directly without a split.
func (a *RouterHandler) GetBestSingleRouteQuote(c echo.Context) error {
	ctx := c.Request().Context()

	tokenOutDenom, tokenIn, err := getValidRoutingParameters(c)
	if err != nil {
		return err
	}

	quote, err := a.RUsecase.GetBestSingleRouteQuote(ctx, tokenIn, tokenOutDenom)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	quote.PrepareResult()

	return c.JSON(http.StatusOK, quote)
}

// GetCandidateRoutes returns the candidate routes for a given tokenIn and tokenOutDenom
func (a *RouterHandler) GetCandidateRoutes(c echo.Context) error {
	ctx := c.Request().Context()

	tokenOutDenom, tokenIn, err := getValidTokenInTokenOutStr(c)
	if err != nil {
		return err
	}

	routes, err := a.RUsecase.GetCandidateRoutes(ctx, tokenIn, tokenOutDenom)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	if err := c.JSON(http.StatusOK, routes); err != nil {
		return err
	}
	return nil
}

// TODO: authentication for the endpoint and enable only in dev mode.
func (a *RouterHandler) StoreRouterStateInFiles(c echo.Context) error {
	ctx := c.Request().Context()

	if err := a.RUsecase.StoreRouterStateFiles(ctx); err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, "Router state stored in files")
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

// getValidRoutingParameters returns the tokenIn and tokenOutDenom from server context if they are valid.
func getValidRoutingParameters(c echo.Context) (string, sdk.Coin, error) {
	tokenOutStr, tokenInStr, err := getValidTokenInTokenOutStr(c)
	if err != nil {
		return "", sdk.Coin{}, err
	}

	matches := coinPattern.FindStringSubmatch(tokenInStr)
	if len(matches) != 3 && len(matches) != 6 {
		return "", sdk.Coin{}, errors.New("tokenIn is invalid - must be in the format amountDenom")
	}

	tokenIn := sdk.Coin{
		Amount: sdk.MustNewDecFromStr(matches[1]).TruncateInt(),
		Denom:  matches[2],
	}

	if err := tokenIn.Validate(); err != nil {
		return "", sdk.Coin{}, c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
	}
	return tokenOutStr, tokenIn, nil
}

func getValidTokenInTokenOutStr(c echo.Context) (tokenOutStr, tokenInStr string, err error) {
	tokenInStr = c.QueryParam("tokenIn")
	tokenOutStr = c.QueryParam("tokenOutDenom")

	if len(tokenInStr) == 0 {
		return "", "", errors.New("tokenIn is required")
	}

	if len(tokenOutStr) == 0 {
		return "", "", errors.New("tokenOutDenom is required")
	}

	return tokenOutStr, tokenInStr, nil
}
