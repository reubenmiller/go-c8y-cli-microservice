package handlers

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/reubenmiller/go-c8y/pkg/c8y"
	"go.uber.org/zap"

	"github.com/labstack/echo"
	"github.com/reubenmiller/c8y-microservice-starter/internal/model"
)

// RegisterHandlers registers the http handlers to the given echo server
func RegisterHandlers(e *echo.Echo) {
	e.Add("POST", "/commands/importevents/run", ExecuteCommandHandler)
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

// ExecuteCommandHandler returns a managed object by its name
func ExecuteCommandHandler(c echo.Context) error {
	cc := c.(*model.RequestContext)

	shouldWait := c.QueryParam("wait")

	inputFile, err := saveFile(c)

	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "Unexpected error processing uploaded file",
			"error":   err.Error(),
		})
	}

	c8yCommand := fmt.Sprintf("cat '%s' | c8y events create --text 'test event' --type 'ms_TestEvent' --select source.id,source.name --output csvheader", inputFile)
	cmd := exec.Command("/bin/sh", "-c", c8yCommand)
	var serviceUser c8y.ServiceUser

	if len(cc.Microservice.Client.ServiceUsers) > 0 {
		serviceUser = cc.Microservice.Client.ServiceUsers[0]
	}

	cmd.Env = []string{
		// Use ms credentials
		"C8Y_HOST=" + cc.Microservice.Client.BaseURL.String(),
		"C8Y_TENANT=" + serviceUser.Tenant,
		"C8Y_USER=" + serviceUser.Username,
		"C8Y_PASSWORD=" + serviceUser.Password,

		// Disable any prompts
		"C8Y_SETTINGS_CI=true",
		"C8Y_SETTINGS_ACTIVITYLOG_ENABLED=/tmp",
		"C8Y_SETTINGS_ACTIVITYLOG_ENABLED=false",
	}

	isAsync := !strings.EqualFold(shouldWait, "true")

	if isAsync {

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

			out, err := cmd.Output()
			if err != nil {
				zap.S().Warnf("Command failed to run. cmd=%s, exit_code=%d, error=%s", c8yCommand, cmd.ProcessState.ExitCode(), err)
			} else {
				zap.S().Warnf("Command was successful. cmd=%s, exit_code=%d", c8yCommand, cmd.ProcessState.ExitCode())
			}

			// post execution result
			event := c8y.NewEventBuilder(cc.Microservice.AgentID, "ms_Command", "Async command finished").
				SetTimestamp(c8y.NewTimestamp()).
				Set("ms_IsAsyncCommand", true).
				Set("ms_Ok", err == nil).
				Set("ms_Command", c8yCommand).
				Set("ms_CommandOutput", string(out))

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
			"async":   isAsync,
			"message": "Started script asynchronously",
		})
	}

	defer os.Remove(inputFile)

	// wait for command to complete
	out, err := cmd.Output()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": fmt.Sprintf("Command failed to run. cmd=%s, exit_code=%d, error=%s", c8yCommand, cmd.ProcessState.ExitCode(), err),
			"details": "",
		})
	}

	zap.S().Debugf("Command output: %s", out)

	return c.JSON(http.StatusOK, echo.Map{
		"ok":      true,
		"async":   isAsync,
		"message": "Command executed successfully",
	})
}
