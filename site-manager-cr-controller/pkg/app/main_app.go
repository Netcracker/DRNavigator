package app

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	_ "github.com/netcracker/drnavigator/site-manager-cr-controller/docs"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/logger"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/model"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// Serve main Server initialize main SM API
func ServeWMainServer(bindAddress string, bindWebhookAddress string, certFile string, keyFile string, crManager service.ICRManager, smConfig *model.SMConfig, errChannel chan error) {
	e := echo.New()
	e.Use(echoprometheus.NewMiddleware("site_manager"))
	e.GET("/", rootGet())
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/health", health(bindWebhookAddress))

	// Authorized group API
	g := e.Group("/sitemanager")
	if envconfig.EnvConfig.FrontHttpAuth {
		g.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, c echo.Context) (bool, error) {
				return smConfig.GetToken() == key, nil
			},
			ErrorHandler: func(err error, c echo.Context) error {
				if _, ok := err.(*middleware.ErrKeyAuthMissing); ok {
					return &echo.HTTPError{
						Code:    http.StatusUnauthorized,
						Message: "You should use Bearer for authorization",
					}
				}
				return &echo.HTTPError{
					Code:    http.StatusForbidden,
					Message: "Bearer is empty or wrong",
				}
			},
		}))
	}
	g.GET("", getServices(crManager))
	g.POST("", processService(crManager))

	if envconfig.EnvConfig.HttpsEnaled {
		errChannel <- e.StartTLS(bindAddress, certFile, keyFile)
	} else {
		errChannel <- e.Start(bindAddress)
	}
}

// rootGet is api GET / func, that is used in main site-manager.
// @Summary      Root request to check SM availability
// @Tags         site-manager
// @Success    200    {string}    "Always return 'Under construction'"
// @Router       / [get]
func rootGet() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.String(http.StatusOK, "Under construction")
	}
}

// health is api GET /health func, that is used in main site-manager.
// @Summary     Health check
// @Tags         site-manager
// @Success    204    "site-manager health up"
// @Router       /health [get]
func health(bindWebhookAddress string) func(c echo.Context) error {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	http_client := &http.Client{Transport: tr}
	return func(c echo.Context) error {
		if bindWebhookAddress != "" {
			resp, err := http_client.Get(fmt.Sprintf("https://%s/health", bindWebhookAddress))
			if err != nil {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("can't check webhook health: %s", err))
			}
			if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
				return c.String(http.StatusInternalServerError, fmt.Sprintf("recived bad status from webhook healthz: %s", resp.Status))
			}
		}
		return c.String(http.StatusNoContent, "")
	}
}

// getServices is api GET /sitemanager func, that is used in main site-manager.
// @Summary     Get the dict of CRs for all services managed by site-manager
// @Tags         site-manager
// @Security BearerAuth
// @Success     200    "CRs dictionary"
// @Failure     401    "Unauthorized user"
// @Failure     403    "Invalid token"
// @Failure     500    "Server error"
// @Router       /sitemanager [get]
func getServices(crManager service.ICRManager) func(c echo.Context) error {
	return func(c echo.Context) error {
		smDict, smErr := crManager.GetAllServices()
		if smErr != nil {
			return c.JSON(smErr.GetStatusCode(), smErr)
		}
		return c.JSON(http.StatusOK, smDict)
	}
}

// processService is api POST /sitemanager func, that is used in main site-manager.
// @Summary			Process service
// @Tags			site-manager
// @Security		BearerAuth
// @Param			sm-request body		model.SMRequest	true	"SM Processing request"
// @Success			200    "Procedure runs"
// @Failure			400    "Wrong data"
// @Failure			401    "Unauthorized user"
// @Failure			403    "Invalid token"
// @Failure			500    "Server error"
// @Router			/sitemanager [post]
func processService(crManager service.ICRManager) func(c echo.Context) error {
	log := logger.SimpleLogger()
	return func(c echo.Context) error {
		smBytes, err := ioutil.ReadAll(c.Request().Body)
		if err != nil {
			log.Errorf("Some problem occurred: %s", err)
			return &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Some problem occurred: %s", err),
			}
		}
		log.Infof("Data was received: %s", smBytes)
		smRequest := model.SMRequest{}
		if err := json.Unmarshal(smBytes, &smRequest); err != nil {
			log.Errorf("Some problem occurred: %s", err)
			return &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "No valid JSON data was received",
			}
		}
		switch smRequest.Procedure {
		case model.ProcedureList:
			listResp, smErr := crManager.GetServicesList()
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(http.StatusOK, listResp)
		case model.ProcedureStatus:
			statusResp, smErr := crManager.GetServiceStatus(smRequest.Service, smRequest.WithDeps)
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(http.StatusOK, statusResp)
		case model.ProcedureActive:
			processResp, smErr := crManager.ProcessService(smRequest.Service, string(smRequest.Procedure), smRequest.NoWait)
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(processResp.GetStatusCode(), processResp)
		case model.ProcedureStandby:
			processResp, smErr := crManager.ProcessService(smRequest.Service, string(smRequest.Procedure), smRequest.NoWait)
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(processResp.GetStatusCode(), processResp)
		case model.ProcedureDisable:
			processResp, smErr := crManager.ProcessService(smRequest.Service, string(smRequest.Procedure), smRequest.NoWait)
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(processResp.GetStatusCode(), processResp)
		default:
			return &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: fmt.Sprintf("You should define procedure from list: %s", model.AllProcedures),
			}
		}
	}
}
