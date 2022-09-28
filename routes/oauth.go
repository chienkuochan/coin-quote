package routes

import (
	"github.com/labstack/echo"
	v1 "quote/api/v1"
)

func SetQuoteInterfaces(e *echo.Echo) {

	e.GET("/api/quotes", v1.GetApiQuotes)

}
