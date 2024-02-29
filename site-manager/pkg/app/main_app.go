package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	_ "github.com/netcracker/drnavigator/site-manager/docs"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	echoSwagger "github.com/swaggo/echo-swagger"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/certwatcher"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var appLog = ctrl.Log.WithName("app-main")

// Serve main Server initialize main SM API
func ServeMainServer(bindAddress string, certDir string, certFile string, keyFile string, crManager service.CRManager, tokenWatcher service.TokenWatcher, errChannel chan error) {
	e := echo.New()
	e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
		Subsystem:  "site_manager",
		Registerer: metrics.Registry,
	}))
	e.GET("/", rootGet())
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.GET("/health", health())

	// Authorized group API
	g := e.Group("/sitemanager")
	if envconfig.EnvConfig.FrontHttpAuth {
		g.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
			Validator: func(key string, c echo.Context) (bool, error) {
				return tokenWatcher.ValidateToken(c.Request().Context(), key)
			},
			ErrorHandler: func(err error, c echo.Context) error {
				if _, ok := err.(*middleware.ErrKeyAuthMissing); ok {
					return &echo.HTTPError{
						Code:    http.StatusUnauthorized,
						Message: "You should use Bearer for authorization",
					}
				}
				if err.Error() != "invalid key" {
					return &echo.HTTPError{
						Code:    http.StatusInternalServerError,
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
		certwatcher, err := certwatcher.New(filepath.Join(certDir, certFile), filepath.Join(certDir, keyFile))
		if err != nil {
			errChannel <- fmt.Errorf("error initializing cert watcher for main API: %s", err)
			return
		}
		go func() {
			errChannel <- certwatcher.Start(context.Background())
		}()
		srv := &http.Server{
			Addr:    bindAddress,
			Handler: e,
			TLSConfig: &tls.Config{
				GetCertificate: certwatcher.GetCertificate,
			},
		}
		errChannel <- srv.ListenAndServeTLS("", "")
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
func health() func(c echo.Context) error {
	return func(c echo.Context) error {
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
func getServices(crManager service.CRManager) func(c echo.Context) error {
	return func(c echo.Context) error {
		smDict, smErr := crManager.GetAllServices(c.Request().Context())
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
func processService(crManager service.CRManager) func(c echo.Context) error {
	return func(c echo.Context) error {
		smBytes, err := io.ReadAll(c.Request().Body)
		if err != nil {
			appLog.Error(err, "Service processing error occurred: %s")
			return &echo.HTTPError{
				Code:    http.StatusInternalServerError,
				Message: fmt.Sprintf("Some problem occurred: %s", err),
			}
		}
		appLog.Info("Data was received", "processing-request", string(smBytes))
		smRequest := model.SMRequest{}
		if err := json.Unmarshal(smBytes, &smRequest); err != nil {
			appLog.Error(err, "Service processing error occurred")
			return &echo.HTTPError{
				Code:    http.StatusBadRequest,
				Message: "No valid JSON data was received",
			}
		}
		switch smRequest.Procedure {
		case model.ProcedureList:
			listResp, smErr := crManager.GetServicesList(c.Request().Context())
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(http.StatusOK, listResp)
		case model.ProcedureStatus:
			statusResp, smErr := crManager.GetServiceStatus(c.Request().Context(), smRequest.Service, smRequest.WithDeps)
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(http.StatusOK, statusResp)
		case model.ProcedureActive:
			processResp, smErr := crManager.ProcessService(c.Request().Context(), smRequest.Service, string(smRequest.Procedure), smRequest.NoWait)
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(processResp.GetStatusCode(), processResp)
		case model.ProcedureStandby:
			processResp, smErr := crManager.ProcessService(c.Request().Context(), smRequest.Service, string(smRequest.Procedure), smRequest.NoWait)
			if smErr != nil {
				return c.JSON(smErr.GetStatusCode(), smErr)
			}
			return c.JSON(processResp.GetStatusCode(), processResp)
		case model.ProcedureDisable:
			processResp, smErr := crManager.ProcessService(c.Request().Context(), smRequest.Service, string(smRequest.Procedure), smRequest.NoWait)
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
