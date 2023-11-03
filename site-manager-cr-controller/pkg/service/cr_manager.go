package service

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	envconfig "github.com/netcracker/drnavigator/site-manager-cr-controller/config"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/logger"
	cr_client "github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/client/cr"
	http_client "github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/client/http"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/model"
	"github.com/netcracker/drnavigator/site-manager-cr-controller/pkg/utils"
	"golang.org/x/exp/maps"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ICRManager is a interface to manage CR objects
type ICRManager interface {
	// GetAllServicesWithSpecifiedVersion returns all sitemanager CRs for specified api version mapped by calculated service name
	GetAllServicesWithSpecifiedVersion(apiVersion string) (*model.SMDictionary, *model.SMError)
	// GetAllServices returns all sitemanager CRs for default api version mapped by calculated service name
	GetAllServices() (*model.SMDictionary, *model.SMError)
	// GetServicesList returns the list of available services
	GetServicesList() (*model.SMListResponse, *model.SMError)
	// GetServiceStatus returns the status of given service
	GetServiceStatus(serviceName *string, withDeps bool) (*model.SMStatusResponse, *model.SMError)
	// ProcessService do given procudedure for given service
	ProcessService(serviceName *string, procedure string, noWait bool) (*model.SMProcedureResponse, *model.SMError)
}

// CRManager is an implementation if ICRManager
type CRManager struct {
	SMConfig       *model.SMConfig
	CRClient       cr_client.ICRClient
	GetHttpClient  http_client.HttpClientInterface
	PostHttpClient http_client.HttpClientInterface
}

// NewCRManager creates new CR manager
func NewCRManager(smConfig *model.SMConfig) (ICRManager, error) {
	crManager := &CRManager{SMConfig: smConfig}
	log := logger.SimpleLogger()

	log.Debugf("Try to initialize http clients for services...")
	tlsConfig := &tls.Config{}
	if caCertEnabled, err := strconv.ParseBool(envconfig.EnvConfig.SMCaCert); err == nil {
		tlsConfig.InsecureSkipVerify = !caCertEnabled
	} else {
		if err := utils.CheckFile(envconfig.EnvConfig.SMCaCert); err != nil {
			return nil, fmt.Errorf("can't initialize http clients: %s", err)
		}
		caCert, err := os.ReadFile(envconfig.EnvConfig.SMCaCert)
		if err != nil {
			return nil, fmt.Errorf("can't initialize http clients: %s", err)
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}
	crManager.GetHttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(envconfig.EnvConfig.GetRequestTimeout) * time.Second,
	}
	crManager.PostHttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: time.Duration(envconfig.EnvConfig.PostRequestTimeout) * time.Second,
	}
	log.Debugf("Http clients are initialized")

	if !crManager.SMConfig.Testing.Enabled {
		// Initialize kube clients for cr for all supported versions
		log.Debugf("Try to initialize kube client for CRs...")
		crClient, err := cr_client.NewCRClient()
		if err != nil {
			return nil, err
		}
		crManager.CRClient = crClient
		log.Debugf("Kube client for CRs was initialized")
	} else {
		log.Debugf("Testing mod is enabled, SM objects will be used from specified sm configuration")
	}
	return crManager, nil
}

// GetAllServicesWithSpecifiedVersion returns all sitemanager CRs for specified api version mapped by calculated service name
func (crm *CRManager) GetAllServicesWithSpecifiedVersion(apiVersion string) (*model.SMDictionary, *model.SMError) {
	if crm.SMConfig != nil && crm.SMConfig.Testing.Enabled {
		return &crm.SMConfig.Testing.SMDict, nil
	}

	log := logger.SimpleLogger()
	crs, err := crm.CRClient.List(apiVersion)
	if err != nil {
		log.Error(err.Error())
		return nil, &model.SMError{Message: err.Error(), IsInternalServerError: true}
	}

	return crm.convertToDict(crs.Items), nil
}

// GetAllServices returns all sitemanager CRs for default api version mapped by calculated service name
func (crm *CRManager) GetAllServices() (*model.SMDictionary, *model.SMError) {
	return crm.GetAllServicesWithSpecifiedVersion(envconfig.EnvConfig.CRVersion)
}

// GetServicesList returns the list of available services
func (crm *CRManager) GetServicesList() (*model.SMListResponse, *model.SMError) {
	smDict, err := crm.GetAllServices()
	if err != nil {
		return nil, err
	}
	return &model.SMListResponse{Services: maps.Keys(smDict.Services)}, nil
}

// GetServiceStatus returns the status of given service
func (crm *CRManager) GetServiceStatus(serviceName *string, withDeps bool) (*model.SMStatusResponse, *model.SMError) {
	smDict, smErr := crm.GetAllServices()
	if smErr != nil {
		return nil, smErr
	}
	_, smErr = crm.getServiceObject(serviceName, smDict)
	if smErr != nil {
		return nil, smErr
	}

	result := &model.SMStatusResponse{Services: map[string]model.SMStatus{}}
	if err := crm.collectServicesStatuses([]string{*serviceName}, smDict, nil, withDeps, result); err != nil {
		return nil, err
	}
	return result, nil
}

// ProcessService do given procudedure for given service
func (crm *CRManager) ProcessService(serviceName *string, procedure string, noWait bool) (*model.SMProcedureResponse, *model.SMError) {
	smDict, smErr := crm.GetAllServices()
	if smErr != nil {
		return nil, smErr
	}
	smObj, smErr := crm.getServiceObject(serviceName, smDict)
	if smErr != nil {
		return nil, smErr
	}
	log := logger.SimpleLogger()
	log.Infof("Service: %s. Set mode %s. serviceEndpoint = %s. No-wait %t", *serviceName, procedure, smObj.Parameters.ServiceEndpoint, noWait)

	processRequest := &model.ServiceProcessRequest{
		Mode:   procedure,
		NoWait: noWait,
	}
	processResponse := &model.ServiceProcessResponse{}
	code, err := http_client.DoPostRequest(crm.PostHttpClient, smObj.Parameters.ServiceEndpoint, processRequest, crm.SMConfig.GetToken(), envconfig.EnvConfig.BackHttpAuth, 3, processResponse)
	if err != nil {
		return nil, &model.SMError{Message: fmt.Sprintf("Processing service error: %s", err), Service: serviceName}
	}
	if code == http.StatusOK {
		return &model.SMProcedureResponse{
			Message:   fmt.Sprintf("Procedure %s is started", procedure),
			Service:   *serviceName,
			Procedure: procedure,
		}, nil
	}
	return &model.SMProcedureResponse{
		Message:   fmt.Sprintf("Procedure %s failed", procedure),
		Service:   *serviceName,
		Procedure: procedure,
		IsFailed:  true,
	}, nil
}

// collectServicesStatuses collects statuses for given services to sm status response
func (crm *CRManager) collectServicesStatuses(services []string, smDict *model.SMDictionary, parentService *string, withDeps bool, resultStatus *model.SMStatusResponse) *model.SMError {
	log := logger.SimpleLogger()
	for _, service := range services {
		serviceObj, err := crm.getServiceObject(&service, smDict)
		if err != nil {
			log.Errorf("Found not exist dependency: %s in %s CR", service, *parentService)
			return &model.SMError{Message: "Dependency defined in CR doesn't exist", Service: err.Service, ProblemCR: parentService}
		}
		if _, found := resultStatus.Services[service]; !found {
			resultStatus.Services[service] = crm.getServiceStatus(serviceObj)
			if withDeps {
				status := resultStatus.Services[service]
				status.Deps = &model.SMStatusDeps{
					After:  serviceObj.After,
					Before: serviceObj.Before,
				}
				resultStatus.Services[service] = status
				if err := crm.collectServicesStatuses(serviceObj.After, smDict, &service, withDeps, resultStatus); err != nil {
					return err
				}
				if err := crm.collectServicesStatuses(serviceObj.Before, smDict, &service, withDeps, resultStatus); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// getServiceStatus returns the status only for specific object (without dependencies)
func (crm *CRManager) getServiceStatus(smObj *model.SMObject) model.SMStatus {
	serviceSMResponse := &model.ServiceSiteManagerResponse{
		Message: "",
		Mode:    "--",
		Status:  "--",
	}

	_, _ = http_client.DoGetRequest(crm.GetHttpClient, smObj.Parameters.ServiceEndpoint, crm.SMConfig.GetToken(), envconfig.EnvConfig.BackHttpAuth, 3, serviceSMResponse)

	serviceHealthResponse := &model.ServiceHealthzResponse{
		Status: "--",
	}
	_, _ = http_client.DoGetRequest(crm.GetHttpClient, smObj.Parameters.HealthzEndpoint, crm.SMConfig.GetToken(), envconfig.EnvConfig.BackHttpAuth, 3, serviceHealthResponse)

	return model.SMStatus{
		Mode:    serviceSMResponse.Mode,
		Status:  serviceSMResponse.Status,
		Message: serviceSMResponse.Message,
		Health:  serviceHealthResponse.Status,
	}
}

// getServiceObject returns the sm object for given service name from sm dictionary
func (crm *CRManager) getServiceObject(serviceName *string, smDict *model.SMDictionary) (*model.SMObject, *model.SMError) {
	if serviceName == nil {
		return nil, &model.SMError{Message: "run-service value should be defined and have String type"}
	}
	if obj, found := smDict.Services[*serviceName]; found {
		log := logger.SimpleLogger()
		log.Infof("Following service will be processed: %s", *serviceName)
		return &obj, nil
	}
	return nil, &model.SMError{Message: "Service doesn't exist", Service: serviceName}
}

// convertToDict converts the list of CRs t SMDict objects
func (crm *CRManager) convertToDict(objList []unstructured.Unstructured) *model.SMDictionary {
	result := &model.SMDictionary{
		Services: map[string]model.SMObject{},
	}
	//TODO apply http and other
	for _, obj := range objList {
		result.Services[cr_client.GetServiceName(&obj)] = model.SMObject{
			CRName:                  obj.GetName(),
			Namespace:               obj.GetNamespace(),
			UID:                     obj.GetUID(),
			Name:                    cr_client.GetServiceName(&obj),
			Module:                  cr_client.GetModule(&obj),
			After:                   cr_client.GetAfter(&obj),
			Before:                  cr_client.GetBefore(&obj),
			Sequence:                cr_client.GetSequence(&obj),
			AllowedStandbyStateList: cr_client.GetAllowedStandbyStateList(&obj),
			Parameters: model.SMObjectParameters{
				ServiceEndpoint: cr_client.GetServiceEndpoint(&obj),
				HealthzEndpoint: cr_client.GetHealthzEndpoint(&obj),
			},
			Timeout: cr_client.GetTimeout(&obj),
			Alias:   cr_client.GetAlias(&obj),
		}
	}
	return result
}
