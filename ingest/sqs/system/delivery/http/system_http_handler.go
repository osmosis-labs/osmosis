package http

import (
	"net/http"
	_ "net/http/pprof"

	"github.com/labstack/echo"
)

// NewSystemHandler will initialize the /debug/ppof resources endpoint
func NewSystemHandler(e *echo.Echo) {

	e.GET("/debug/pprof/*", echo.WrapHandler(http.DefaultServeMux))

	// // Register pprof handlers on "/debug/pprof"
	// e.GET("/debug/pprof/*", echo.WrapHandler(http.HandlerFunc(pprof.Index)))
}
