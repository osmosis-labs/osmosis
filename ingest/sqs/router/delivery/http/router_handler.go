package http

import (
	"net/http"
	"regexp"

	"github.com/labstack/echo"
	"github.com/sirupsen/logrus"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain"
)

// ResponseError represent the response error struct
type ResponseError struct {
	Message string `json:"message"`
}

// RouterHandler  represent the httphandler for the router
type RouterHandler struct {
	RUsecase domain.RouterUsecase
}

// Define a regular expression pattern to match sdk.Coin where the first part is the amount and second is the denom name
var coinPattern = regexp.MustCompile(`([0-9]+)([a-z]+)`)

// NewRouterHandler will initialize the pools/ resources endpoint
func NewRouterHandler(e *echo.Echo, us domain.RouterUsecase) {
	handler := &RouterHandler{
		RUsecase: us,
	}
	e.GET("/quote", handler.GetOptimalQuote)
}

// GetOptimalQuote will determine the optimal quote for a given tokenIn and tokenOutDenom
// Return the optimal quote.
func (a *RouterHandler) GetOptimalQuote(c echo.Context) error {
	ctx := c.Request().Context()

	tokenInStr := c.QueryParam("tokenIn")
	tokenOutDenom := c.QueryParam("tokenOutDenom")

	if len(tokenInStr) == 0 {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "tokenIn is required"})
	}

	if len(tokenOutDenom) == 0 {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "tokenOutDenom is required"})
	}

	matches := coinPattern.FindStringSubmatch(tokenInStr)
	if len(matches) != 3 {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: "tokenIn is invalid - must be in the format amountDenom"})
	}

	tokenIn := sdk.Coin{
		Amount: sdk.MustNewDecFromStr(matches[1]).TruncateInt(),
		Denom:  matches[2],
	}

	if err := tokenIn.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, ResponseError{Message: err.Error()})
	}

	quote, err := a.RUsecase.GetOptimalQuote(ctx, tokenIn, tokenOutDenom)
	if err != nil {
		return c.JSON(getStatusCode(err), ResponseError{Message: err.Error()})
	}

	return c.JSON(http.StatusOK, quote)
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
