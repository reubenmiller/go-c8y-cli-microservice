package model

import (
	"github.com/labstack/echo"
	"github.com/reubenmiller/go-c8y/pkg/microservice"
)

// RequestContext Custom context used for each handler in order to provide access to the C8Y Client and configuration
type RequestContext struct {
	echo.Context

	Microservice *microservice.Microservice
}
