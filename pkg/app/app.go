package app

import (
	"log"

	"github.com/reubenmiller/go-c8y-cli-microservice/pkg/tasks"

	"github.com/labstack/echo"
	"github.com/reubenmiller/go-c8y-cli-microservice/internal/model"
	"github.com/reubenmiller/go-c8y-cli-microservice/pkg/handlers"
	"github.com/reubenmiller/go-c8y/pkg/microservice"
	"go.uber.org/zap"
)

// App represents the http server and c8y microservice application
type App struct {
	echoServer      *echo.Echo
	c8ymicroservice *microservice.Microservice
}

// NewApp initializes the microservice with default configuration and registers the microservice
func NewApp() *App {
	app := &App{}
	log.Printf("Application information: Version %s, branch %s, commit %s, buildTime %s", Version, Branch, Commit, BuildTime)

	opts := microservice.Options{}
	opts.AgentInformation = microservice.AgentInformation{
		SerialNumber: Commit,
		Revision:     Version,
		BuildTime:    BuildTime,
	}

	c8ymicroservice := microservice.NewDefaultMicroservice(opts)

	// Set app defaults before registering the microservice
	c8ymicroservice.Config.SetDefault("server.port", "80")

	c8ymicroservice.RegisterMicroserviceAgent()
	app.c8ymicroservice = c8ymicroservice

	app.c8ymicroservice.Scheduler.AddFunc(
		"@every 30m",
		tasks.ExampleHeartbeatTask(app.c8ymicroservice),
	)

	return app
}

// Run starts the microservice
func (a *App) Run() {
	application := a.c8ymicroservice
	application.Scheduler.Start()

	if a.echoServer == nil {
		addr := ":" + application.Config.GetString("server.port")
		zap.S().Infof("starting http server on %s", addr)

		a.echoServer = echo.New()
		setDefaultContextHandler(a.echoServer, a.c8ymicroservice)

		a.setRouters()

		a.echoServer.Logger.Fatal(a.echoServer.Start(addr))
	}
}

func setDefaultContextHandler(e *echo.Echo, c8yms *microservice.Microservice) {
	// Add Custom Context
	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &model.RequestContext{
				Context:      c,
				Microservice: c8yms,
			}
			return h(cc)
		}
	})
}

func (a *App) setRouters() {
	server := a.echoServer

	/*
	 ** Routes
	 */
	// TODO: Add routes here
	handlers.RegisterHandlers(server)

	/*
	 ** Health endpoints
	 */
	a.c8ymicroservice.AddHealthEndpointHandlers(server)
}
