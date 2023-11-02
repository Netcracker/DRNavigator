package app

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/logger"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/service"
	admissionv1 "k8s.io/api/admission/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Serve webhook Server initialize webhook API
func ServeWebhookServer(bindAddress string, certFile string, keyFile string, validator service.IValidator, converter service.IConverter, errChannel chan error) {
	e := echo.New()
	e.Use(echoprometheus.NewMiddleware("sm_cr_controller"))
	e.GET("/metrics", echoprometheus.NewHandler())
	e.GET("/health", webhookHealth())
	e.POST("/validate", validate(validator))
	e.POST("/convert", convert(converter))
	errChannel <- e.StartTLS(bindAddress, certFile, keyFile)
}

// health is api GET /health func, that is used in webhook
func webhookHealth() func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.String(http.StatusNoContent, "")
	}
}

// validate is api POST /validate func, that is used in webhook
func validate(validator service.IValidator) func(c echo.Context) error {
	return func(c echo.Context) error {
		log := logger.SimpleLogger()

		input := admissionv1.AdmissionReview{}
		if err := json.NewDecoder(c.Request().Body).Decode(&input); err != nil {
			msg := fmt.Sprintf("Can't parse admission review object: %s", err)
			log.Errorf(msg)
			return c.String(http.StatusInternalServerError, msg)
		}

		log.Debugf("Initial object from API for validating:\n%s", string(input.Request.Object.Raw))

		cr := unstructured.Unstructured{}
		if err := cr.UnmarshalJSON(input.Request.Object.Raw); err != nil {
			msg := fmt.Sprintf("Can't parse cr object: %s", err)
			log.Errorf(msg)
			return c.String(http.StatusInternalServerError, msg)
		}
		allowed, message, err := validator.Validate(&cr)

		if err != nil {
			log.Errorf(err.Error())
			return c.String(http.StatusInternalServerError, err.Error())
		}
		if !allowed {
			log.Debugf("CR validation fails: %s", message)
		}

		output := admissionv1.AdmissionReview{
			TypeMeta: input.TypeMeta,
			Response: &admissionv1.AdmissionResponse{
				UID:     input.Request.UID,
				Allowed: allowed,
				Result: &metav1.Status{
					Message: message,
				},
			},
		}
		return c.JSON(http.StatusOK, output)
	}
}

// convert is api POST /convert func, that is used in webhook
func convert(converter service.IConverter) func(c echo.Context) error {
	return func(c echo.Context) error {
		log := logger.SimpleLogger()
		input := apiextensionsv1.ConversionReview{}
		if err := json.NewDecoder(c.Request().Body).Decode(&input); err != nil {
			msg := fmt.Sprintf("Can't parse conversion review object: %s", err)
			log.Error(msg)
			return c.String(http.StatusInternalServerError, msg)
		}

		log.Debugf("CR conversation is started.")
		var convertedObjects []runtime.RawExtension

		for _, obj := range input.Request.Objects {
			cr := unstructured.Unstructured{}
			if err := cr.UnmarshalJSON(obj.Raw); err != nil {
				msg := fmt.Sprintf("Can't parse unstructured object: %s", err)
				log.Error(msg)
				return c.String(http.StatusInternalServerError, msg)
			}
			log.Debugf("Initial spec:%s", cr)
			convertedCR, err := converter.Convert(&cr, input.Request.DesiredAPIVersion)

			if err != nil {
				return c.String(http.StatusInternalServerError, err.Error())
			}

			log.Debugf("Modified spec: %s", convertedCR)
			convertedObjects = append(convertedObjects, runtime.RawExtension{Object: convertedCR})
		}

		output := apiextensionsv1.ConversionReview{
			TypeMeta: input.TypeMeta,
			Response: &apiextensionsv1.ConversionResponse{
				UID: input.Request.UID,
				Result: metav1.Status{
					Status: "Success",
				},
				ConvertedObjects: convertedObjects,
			},
		}
		return c.JSON(http.StatusOK, output)
	}
}
