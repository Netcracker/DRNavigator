package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	crv3 "github.com/netcracker/drnavigator/site-manager/api/legacy/v3"
	qubershiporgv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	cr_client "github.com/netcracker/drnavigator/site-manager/pkg/client/cr"
	http_client "github.com/netcracker/drnavigator/site-manager/pkg/client/http"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/utils"
	"golang.org/x/exp/maps"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var crManagerLog = ctrl.Log.WithName("cr-manager")

// CRManager is a interface to manage CR objects
type CRManager interface {
	// GetAllServices returns all sitemanager CRs for default api version mapped by calculated service name
	GetAllServices(ctx context.Context) (*model.SMDictionary, *model.SMError)
	// GetServicesList returns the list of available services
	GetServicesList(ctx context.Context) (*model.SMListResponse, *model.SMError)
	// GetServiceStatus returns the status of given service
	GetServiceStatus(ctx context.Context, serviceName *string, withDeps bool) (*model.SMStatusResponse, *model.SMError)
	// ProcessService do given procudedure for given service
	ProcessService(ctx context.Context, serviceName *string, procedure string, noWait bool) (*model.SMProcedureResponse, *model.SMError)
}

// CRManagerImpl is an implementation if CRManager
type CRManagerImpl struct {
	SMConfig       *model.SMConfig
	CRClient       cr_client.CRClient
	TokenWatcher   TokenWatcher
	GetHttpClient  http_client.HttpClientInterface
	PostHttpClient http_client.HttpClientInterface
}

// NewCRManager creates new CR manager
func NewCRManager(smConfig *model.SMConfig, crClient cr_client.CRClient, tokenWatcker TokenWatcher) (CRManager, error) {
	crManager := &CRManagerImpl{SMConfig: smConfig, TokenWatcher: tokenWatcker}

	crManagerLog.V(1).Info("Try to initialize http clients for services...")
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
			TLSClientConfig:   tlsConfig,
			DisableKeepAlives: true,
		},
		Timeout: time.Duration(envconfig.EnvConfig.GetRequestTimeout) * time.Second,
	}

	crManager.PostHttpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   tlsConfig,
			DisableKeepAlives: true,
		},
		Timeout: time.Duration(envconfig.EnvConfig.PostRequestTimeout) * time.Second,
	}
	crManagerLog.V(1).Info("Http clients are initialized")

	if !crManager.SMConfig.Testing.Enabled {
		crManager.CRClient = crClient
	} else {
		crManagerLog.V(1).Info("Testing mod is enabled, SM objects will be used from specified sm configuration")
	}
	return crManager, nil
}

// GetAllServices returns all sitemanager CRs for default api version mapped by calculated service name
func (crm *CRManagerImpl) GetAllServices(ctx context.Context) (*model.SMDictionary, *model.SMError) {
	if crm.SMConfig != nil && crm.SMConfig.Testing.Enabled {
		return &crm.SMConfig.Testing.SMDict, nil
	}

	legacyCRs, err := crm.CRClient.ListLegacy(ctx, &client.ListOptions{})
	if err != nil {
		crManagerLog.Error(err, fmt.Sprintf("can't get sitemanager legacy objects group=%s, version=%s, kind=%s",
			envconfig.EnvConfig.CRGroup, crv3.CRVersion, envconfig.EnvConfig.CRKindList))
		return nil, &model.SMError{Message: err.Error(), IsInternalServerError: true}
	}

	crs := &qubershiporgv3.SiteManagerList{}
	err = crm.CRClient.List(ctx, crs)
	if err != nil {
		crManagerLog.Error(err, "can't list SiteManager resources")
		return nil, &model.SMError{Message: err.Error(), IsInternalServerError: true}
	}

	return crm.convertToDict(legacyCRs, crs), nil
}

// GetServicesList returns the list of available services
func (crm *CRManagerImpl) GetServicesList(ctx context.Context) (*model.SMListResponse, *model.SMError) {
	smDict, err := crm.GetAllServices(ctx)
	if err != nil {
		return nil, err
	}
	return &model.SMListResponse{Services: maps.Keys(smDict.Services)}, nil
}

// GetServiceStatus returns the status of given service
func (crm *CRManagerImpl) GetServiceStatus(ctx context.Context, serviceName *string, withDeps bool) (*model.SMStatusResponse, *model.SMError) {
	smDict, smErr := crm.GetAllServices(ctx)
	if smErr != nil {
		return nil, smErr
	}
	_, smErr = crm.getServiceObject(serviceName, smDict, true)
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
func (crm *CRManagerImpl) ProcessService(ctx context.Context, serviceName *string, procedure string, noWait bool) (*model.SMProcedureResponse, *model.SMError) {
	smDict, smErr := crm.GetAllServices(ctx)
	if smErr != nil {
		return nil, smErr
	}
	smObj, smErr := crm.getServiceObject(serviceName, smDict, false)
	if smErr != nil {
		return nil, smErr
	}
	crManagerLog.Info("Process service", "service-name", *serviceName, "mode", procedure, "service-endpoint", smObj.Parameters.ServiceEndpoint, "no-wait", noWait)

	processRequest := &model.ServiceProcessRequest{
		Mode:   procedure,
		NoWait: noWait,
	}
	processResponse := &model.ServiceProcessResponse{}
	code, err := http_client.DoPostRequest(crm.PostHttpClient, smObj.Parameters.ServiceEndpoint, processRequest, crm.TokenWatcher.GetToken(), envconfig.EnvConfig.BackHttpAuth, 3, processResponse)
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
func (crm *CRManagerImpl) collectServicesStatuses(services []string, smDict *model.SMDictionary, parentService *string, withDeps bool, resultStatus *model.SMStatusResponse) *model.SMError {
	for _, service := range services {
		serviceObj, err := crm.getServiceObject(&service, smDict, false)
		if err != nil {
			crManagerLog.Error(nil, "Found not exist dependency", "dep", service, "problem-cr", *parentService)
			return &model.SMError{Message: "Dependency defined in CR doesn't exist", Service: err.Service, ProblemCR: parentService}
		}
		if _, found := resultStatus.Services[service]; !found {
			resultStatus.Services[service], err = crm.getServiceStatus(serviceObj)
			if err != nil {
				return err
			}
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
func (crm *CRManagerImpl) getServiceStatus(smObj *model.SMObject) (model.SMStatus, *model.SMError) {
	serviceSMResponse := &model.ServiceSiteManagerResponse{
		Message: "",
		Mode:    "--",
		Status:  "--",
	}

	_, err := http_client.DoGetRequest(crm.GetHttpClient, smObj.Parameters.ServiceEndpoint, crm.TokenWatcher.GetToken(), envconfig.EnvConfig.BackHttpAuth, 3, serviceSMResponse)
	if err != nil {
		return model.SMStatus{Message: "", Mode: "--", Status: "--", Health: "--"},
			&model.SMError{Message: fmt.Sprintf("Service request failed, error: %s", err), Service: &smObj.Name}
	}

	serviceHealthResponse := &model.ServiceHealthzResponse{
		Status: "--",
	}
	_, err = http_client.DoGetRequest(crm.GetHttpClient, smObj.Parameters.HealthzEndpoint, crm.TokenWatcher.GetToken(), envconfig.EnvConfig.BackHttpAuth, 3, serviceHealthResponse)
	if err != nil {
		return model.SMStatus{Message: "", Mode: "--", Status: "--", Health: "--"},
			&model.SMError{Message: fmt.Sprintf("Service request failed, error: %s", err), Service: &smObj.Name}
	}

	return model.SMStatus{
		Mode:    serviceSMResponse.Mode,
		Status:  serviceSMResponse.Status,
		Message: serviceSMResponse.Message,
		Health:  serviceHealthResponse.Status,
	}, nil
}

// getServiceObject returns the sm object for given service name from sm dictionary
func (crm *CRManagerImpl) getServiceObject(serviceName *string, smDict *model.SMDictionary, silent bool) (*model.SMObject, *model.SMError) {
	if serviceName == nil {
		return nil, &model.SMError{Message: "run-service value should be defined and have String type"}
	}
	if obj, found := smDict.Services[*serviceName]; found {
		if !silent {
			crManagerLog.Info("Following service will be processed", "service-name", *serviceName)
		}
		return &obj, nil
	}
	return nil, &model.SMError{Message: "Service doesn't exist", Service: serviceName}
}

// convertToDict converts the list of CRs t SMDict objects
func (crm *CRManagerImpl) convertToDict(legacyList *crv3.CRList, list *qubershiporgv3.SiteManagerList) *model.SMDictionary {
	result := &model.SMDictionary{
		Services: map[string]model.SMObject{},
	}
	for _, obj := range legacyList.Items {
		smObj := model.SMObject{
			CRName:                  obj.GetName(),
			Namespace:               obj.GetNamespace(),
			UID:                     obj.GetUID(),
			Name:                    obj.GetServiceName(),
			Module:                  obj.Spec.SiteManager.Module,
			After:                   obj.Spec.SiteManager.After,
			Before:                  obj.Spec.SiteManager.Before,
			Sequence:                obj.Spec.SiteManager.Sequence,
			AllowedStandbyStateList: obj.Spec.SiteManager.AllowedStandbyStateList,
			Parameters: model.SMObjectParameters{
				ServiceEndpoint: obj.Spec.SiteManager.Parameters.ServiceEndpoint,
				HealthzEndpoint: obj.Spec.SiteManager.Parameters.HealthzEndpoint,
			},
			Timeout: obj.Spec.SiteManager.Timeout,
			Alias:   obj.Spec.SiteManager.Alias,
		}
		applyDefaults(&smObj)
		result.Services[obj.GetServiceName()] = smObj
	}

	for _, obj := range list.Items {
		smObj := model.SMObject{
			CRName:                  obj.GetName(),
			Namespace:               obj.GetNamespace(),
			UID:                     obj.GetUID(),
			Name:                    obj.GetServiceName(),
			Module:                  obj.Spec.SiteManager.Module,
			After:                   obj.Spec.SiteManager.After,
			Before:                  obj.Spec.SiteManager.Before,
			Sequence:                obj.Spec.SiteManager.Sequence,
			AllowedStandbyStateList: obj.Spec.SiteManager.AllowedStandbyStateList,
			Parameters: model.SMObjectParameters{
				ServiceEndpoint: obj.Spec.SiteManager.Parameters.ServiceEndpoint,
				HealthzEndpoint: obj.Spec.SiteManager.Parameters.HealthzEndpoint,
			},
			Timeout: obj.Spec.SiteManager.Timeout,
			Alias:   obj.Spec.SiteManager.Alias,
		}
		applyDefaults(&smObj)
		if prev, ok := result.Services[obj.GetServiceName()]; ok {
			crManagerLog.Info(
				"Legacy resource is shadowed by new resource",
				"service-name", obj.GetServiceName(),
				"legacy-resource", fmt.Sprintf("%s/%s", prev.Namespace, prev.CRName),
				"new-resource", fmt.Sprintf("%s/%s", smObj.Namespace, smObj.CRName),
			)
		}
		result.Services[obj.GetServiceName()] = smObj
	}
	return result
}

// applyDefaults applies default values to specified obj
func applyDefaults(obj *model.SMObject) {
	if obj.After == nil {
		obj.After = []string{}
	}
	if obj.Before == nil {
		obj.Before = []string{}
	}
	if len(obj.Module) == 0 {
		obj.Module = "stateful"
	}
	if len(obj.Sequence) == 0 {
		obj.Sequence = []string{"standby", "active"}
	}
	if len(obj.AllowedStandbyStateList) == 0 {
		obj.AllowedStandbyStateList = []string{"up"}
	}
	obj.Parameters.ServiceEndpoint = applyHttpScheme(obj.Parameters.ServiceEndpoint)
	obj.Parameters.HealthzEndpoint = applyHttpScheme(obj.Parameters.HealthzEndpoint)
}

// applyHttpScheme apply defaunt http scheme to endpoint if it's not already presended
func applyHttpScheme(endpoint string) string {
	if len(endpoint) == 0 || strings.HasPrefix(endpoint, "http://") || strings.HasPrefix(endpoint, "https://") {
		return endpoint
	}
	return fmt.Sprintf("%s%s", envconfig.EnvConfig.HttpScheme, endpoint)
}
