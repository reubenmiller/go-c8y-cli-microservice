package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"go.uber.org/zap"

	"github.com/labstack/echo"
	"github.com/reubenmiller/go-c8y-cli-microservice/internal/model"
	"github.com/reubenmiller/go-c8y-cli-microservice/pkg/c8ycli"
)

// RegisterHandlers registers the http handlers to the given echo server
func RegisterHandlers(e *echo.Echo) {
	e.Add("POST", "/commands/importevents/async", ImportEventsFactory(false))
	e.Add("POST", "/commands/importevents/sync", ImportEventsFactory(true))
	e.Add("POST", "/commands/debug", CommandHandler)
}

func saveFile(c echo.Context) (string, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return "", err
	}

	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Destination
	dst, err := ioutil.TempFile("", "upload-"+file.Filename)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy
	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}

	return dst.Name(), nil
}

type (
	CLICommand struct {
		Command string `json:"command"`
	}
)

func CommandHandler(c echo.Context) error {
	cc := c.(*model.RequestContext)

	cliCommand := new(CLICommand)
	if err := c.Bind(cliCommand); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	creds := cc.Microservice.WithServiceUserCredentials()
	executor := &c8ycli.Executor{
		Command: cliCommand.Command,
		Options: &c8ycli.CLIOptions{
			Host:         cc.Microservice.Client.BaseURL.String(),
			Tenant:       creds.Tenant,
			Username:     creds.Username,
			Password:     creds.Password,
			EnableCreate: true,
			EnableUpdate: true,
			EnableDelete: false,
		},
	}
	result, _ := executor.Execute(true)

	return c.JSON(http.StatusOK, echo.Map{
		"message":    "Command finished",
		"command":    cliCommand.Command,
		"stdout":     string(result.Stdout),
		"stderr":     string(result.Stderr),
		"exitCode":   result.ExitCode,
		"durationMS": result.Cmd.ProcessState.UserTime().Milliseconds(),
	})
}

// ExecuteCommandHandler returns a managed object by its name
func ImportEventsFactory(wait bool) func(echo.Context) error {
	return func(c echo.Context) error {
		return ExecuteImportEvents(c, wait)
	}
}
func ExecuteImportEvents(c echo.Context, wait bool) error {
	cc := c.(*model.RequestContext)

	inputFile, err := saveFile(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unexpected error processing uploaded file",
			"error":   err.Error(),
		})
	}

	c8yCommand := fmt.Sprintf("cat '%s' | c8y events create --text 'test event' --type 'ms_TestEvent' --select source.id,source.name --output csvheader", inputFile)

	creds := cc.Microservice.WithServiceUserCredentials()
	executor := &c8ycli.Executor{
		Command: c8yCommand,
		Options: &c8ycli.CLIOptions{
			Host:         cc.Microservice.Client.BaseURL.String(),
			Tenant:       creds.Tenant,
			Username:     creds.Username,
			Password:     creds.Password,
			EnableCreate: true,
			EnableUpdate: false,
			EnableDelete: false,
		},
	}

	if !wait {

		// Log the requests in c8y
		cc.Microservice.Client.Event.Create(
			cc.Microservice.WithServiceUser(),
			c8y.NewEventBuilder(cc.Microservice.AgentID, "ms_Command", "Executing async command").
				SetTimestamp(c8y.NewTimestamp()).
				Set("ms_IsAsyncCommand", true).
				Set("ms_Command", c8yCommand),
		)

		// Don't wait for command to complete
		go func() {
			defer os.Remove(inputFile)

			result, err := executor.Execute(true)

			if err != nil {
				zap.S().Warnf("Command failed to run. cmd=%s, exit_code=%d, error=%s", c8yCommand, result.ExitCode, err)
			} else {
				zap.S().Warnf("Command was successful. cmd=%s, exit_code=%d", c8yCommand, result.ExitCode)
			}

			// post execution result
			event := c8y.NewEventBuilder(cc.Microservice.AgentID, "ms_Command", "Async command finished").
				SetTimestamp(c8y.NewTimestamp()).
				Set("ms_Command", c8yCommand).
				Set("ms_Command_ExitCode", result.ExitCode).
				Set("ms_command_Stdout", string(result.Stdout)).
				Set("ms_command_Stderr", string(result.Stderr))

			if err != nil {
				event.Set("ms_Err", err.Error())
			}
			cc.Microservice.Client.Event.Create(
				cc.Microservice.WithServiceUser(),
				event,
			)
		}()
		return c.JSON(http.StatusOK, echo.Map{
			"ok":      true,
			"message": "Started command",
			"async":   !wait,
		})
	}

	defer os.Remove(inputFile)

	// wait for command to complete
	result, err := executor.Execute(true)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": fmt.Sprintf("Command failed to run. cmd=%s, exit_code=%d, error=%s", c8yCommand, result.ExitCode, err),
			"details": "",
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"ok":      true,
		"message": "Command executed successfully",
		"async":   !wait,
		"stdout":  string(result.Stdout),
		"stderr":  string(result.Stderr),
	})
}
