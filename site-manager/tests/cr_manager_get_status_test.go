package test

import (
	"context"
	"testing"

	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	"github.com/netcracker/drnavigator/site-manager/pkg/model"
	"github.com/netcracker/drnavigator/site-manager/pkg/service"
	mock "github.com/netcracker/drnavigator/site-manager/tests/mock"
	"github.com/stretchr/testify/require"
)

// serviceD--------------------------|
//    V                              V
// serviceC -----> serviceB ----> serviceA
//   |                               ^
//   |_______________________________|

var (
	parameters = model.SMObjectParameters{
		ServiceEndpoint: "http://stub-endpoint8080/sitemanager",
		HealthzEndpoint: "http://stub-endpoint8080/health",
	}

	serviceA = "serviceA"
	serviceB = "serviceB"
	serviceC = "serviceC"
	serviceD = "serviceD"
	smConfig = model.SMConfig{
		Testing: model.SMConfigTesting{
			Enabled: true,
			SMDict: model.SMDictionary{
				Services: map[string]model.SMObject{
					serviceA: {
						After:      []string{},
						Before:     []string{},
						Parameters: parameters,
					},
					serviceB: {
						After:      []string{},
						Before:     []string{serviceA},
						Parameters: parameters,
					},
					serviceC: {
						After:      []string{serviceB},
						Before:     []string{serviceA},
						Parameters: parameters,
					},
					serviceD: {
						After:      []string{serviceA, serviceC},
						Before:     []string{},
						Parameters: parameters,
					},
				},
			},
		},
	}

	expectedStatusA = model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{},
			Before: []string{},
		},
	}
	expectedStatusB = model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{},
			Before: []string{"serviceA"},
		},
	}
	expectedStatusC = model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{"serviceB"},
			Before: []string{"serviceA"},
		},
	}
	expectedStatusD = model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{"serviceA", "serviceC"},
			Before: []string{},
		},
	}
)

func TestCRManager_StatusWithDeps(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	httpClientMock := mock.HttpClientMock{
		ServiceStatus: model.ServiceSiteManagerResponse{Mode: "active", Status: "done"},
		ServiceHealth: model.ServiceHealthzResponse{Status: "up"},
	}
	tokenWatcher, err := service.NewTokenWatcher(&smConfig, nil, "")
	assert.Nil(err, "Can't create token watcher")

	crManager := service.CRManagerImpl{
		SMConfig:      &smConfig,
		GetHttpClient: &httpClientMock,
		TokenWatcher: tokenWatcher,
	}

	serviceAStatus, err := crManager.GetServiceStatus(context.Background(), &serviceA, true)
	assert.Nil(err, "Can't get status for %s", serviceA)
	assert.Equal(&model.SMStatusResponse{
		Services: map[string]model.SMStatus{
			serviceA: expectedStatusA,
		},
	}, serviceAStatus, "status for serviceA is not desired")

	serviceBStatus, err := crManager.GetServiceStatus(context.Background(), &serviceB, true)
	assert.Nil(err, "Can't get status for %s", serviceB)
	assert.Equal(&model.SMStatusResponse{
		Services: map[string]model.SMStatus{
			serviceA: expectedStatusA,
			serviceB: expectedStatusB,
		},
	}, serviceBStatus, "status for serviceB is not desired")

	serviceCStatus, err := crManager.GetServiceStatus(context.Background(), &serviceC, true)
	assert.Nil(err, "Can't get status for %s", serviceC)
	assert.Equal(&model.SMStatusResponse{
		Services: map[string]model.SMStatus{
			serviceA: expectedStatusA,
			serviceB: expectedStatusB,
			serviceC: expectedStatusC,
		},
	}, serviceCStatus, "status for serviceC is not desired")

	serviceDStatus, err := crManager.GetServiceStatus(context.Background(), &serviceD, true)
	assert.Nil(err, "Can't get status for %s", serviceD)
	assert.Equal(&model.SMStatusResponse{
		Services: map[string]model.SMStatus{
			serviceA: expectedStatusA,
			serviceB: expectedStatusB,
			serviceC: expectedStatusC,
			serviceD: expectedStatusD,
		},
	}, serviceDStatus, "status for serviceD is not desired")
}

func TestCRManager_StatusWithoutDeps(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)
	httpClientMock := mock.HttpClientMock{
		ServiceStatus: model.ServiceSiteManagerResponse{Mode: "active", Status: "done"},
		ServiceHealth: model.ServiceHealthzResponse{Status: "up"},
	}
	tokenWatcher, err := service.NewTokenWatcher(&smConfig, nil, "")
	assert.Nil(err, "Can't create token watcher")

	crManager := service.CRManagerImpl{
		SMConfig:      &smConfig,
		GetHttpClient: &httpClientMock,
		TokenWatcher: tokenWatcher,
	}

	serviceAStatus, err := crManager.GetServiceStatus(context.Background(), &serviceA, false)
	assert.Nil(err, "Can't get status for %s", serviceA)
	assert.Equal(1, len(serviceAStatus.Services), "only one service should be in status without deps")
	assert.Contains(serviceAStatus.Services, serviceA, "serviceA should be specified in status")
	assert.Nil(serviceAStatus.Services[serviceA].Deps, "deps should not be in status")

	serviceBStatus, err := crManager.GetServiceStatus(context.Background(), &serviceB, false)
	assert.Nil(err, "Can't get status for %s", serviceB)
	assert.Equal(1, len(serviceBStatus.Services), "only one service should be in status without deps")
	assert.Contains(serviceBStatus.Services, serviceB, "serviceB should be specified in status")
	assert.Nil(serviceBStatus.Services[serviceB].Deps, "deps should not be in status")

	serviceCStatus, err := crManager.GetServiceStatus(context.Background(), &serviceC, false)
	assert.Nil(err, "Can't get status for %s", serviceC)
	assert.Equal(1, len(serviceCStatus.Services), "only one service should be in status without deps")
	assert.Contains(serviceCStatus.Services, serviceC, "serviceC should be specified in status")
	assert.Nil(serviceCStatus.Services[serviceC].Deps, "deps should not be in status")

	serviceDStatus, err := crManager.GetServiceStatus(context.Background(), &serviceD, false)
	assert.Nil(err, "Can't get status for %s", serviceD)
	assert.Equal(1, len(serviceDStatus.Services), "only one service should be in status without deps")
	assert.Contains(serviceDStatus.Services, serviceD, "serviceD should be specified in status")
	assert.Nil(serviceDStatus.Services[serviceD].Deps, "deps should not be in status")
}

func TestCRManager_NotExistDeps(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)

	httpClientMock := mock.HttpClientMock{
		ServiceStatus: model.ServiceSiteManagerResponse{Mode: "active", Status: "done"},
		ServiceHealth: model.ServiceHealthzResponse{Status: "up"},
	}
	tokenWatcher, err := service.NewTokenWatcher(&smConfig, nil, "")
	assert.Nil(err, "Can't create token watcher")

	crManager := service.CRManagerImpl{
		SMConfig:      &smConfig,
		GetHttpClient: &httpClientMock,
		TokenWatcher: tokenWatcher,
	}

	// Add service with wrong dependency
	serviceE := "serviceE"
	serviceF := "serviceF"
	NotExistDep := "not-exist"
	smConfig.Testing.SMDict.Services[serviceE] = model.SMObject{
		After:      []string{},
		Before:     []string{serviceF},
		Parameters: parameters,
	}
	smConfig.Testing.SMDict.Services[serviceF] = model.SMObject{
		After:      []string{},
		Before:     []string{NotExistDep},
		Parameters: parameters,
	}

	// Check, that wrong deps don't affect other services
	serviceDStatus, err := crManager.GetServiceStatus(context.Background(), &serviceD, true)
	assert.Nil(err, "Can't get status for %s", serviceD)
	assert.Equal(&model.SMStatusResponse{
		Services: map[string]model.SMStatus{
			serviceA: expectedStatusA,
			serviceB: expectedStatusB,
			serviceC: expectedStatusC,
			serviceD: expectedStatusD,
		},
	}, serviceDStatus, "status for serviceD is not desired")

	// Check, that status is returned without deps
	serviceEStatus, err := crManager.GetServiceStatus(context.Background(), &serviceE, false)
	assert.Nil(err, "Can't get status for %s", serviceE)
	assert.Equal(1, len(serviceEStatus.Services), "only one service should be in status without deps")
	assert.Contains(serviceEStatus.Services, serviceE, "serviceE should be specified in status")
	assert.Nil(serviceEStatus.Services[serviceE].Deps, "deps should not be in status")

	serviceFStatus, err := crManager.GetServiceStatus(context.Background(), &serviceF, false)
	assert.Nil(err, "Can't get status for %s", serviceF)
	assert.Equal(1, len(serviceFStatus.Services), "only one service should be in status without deps")
	assert.Contains(serviceFStatus.Services, serviceF, "serviceF should be specified in status")
	assert.Nil(serviceFStatus.Services[serviceF].Deps, "deps should not be in status")

	// Check, that exception throws with deps
	_, err = crManager.GetServiceStatus(context.Background(), &serviceE, true)
	assert.Equal(&model.SMError{
		Message:   "Dependency defined in CR doesn't exist",
		Service:   &NotExistDep,
		ProblemCR: &serviceF,
	}, err, "wrong dependency error should be returned")
}

func TestCRManager_CRCycles(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)

	httpClientMock := mock.HttpClientMock{
		ServiceStatus: model.ServiceSiteManagerResponse{Mode: "active", Status: "done"},
		ServiceHealth: model.ServiceHealthzResponse{Status: "up"},
	}
	tokenWatcher, err := service.NewTokenWatcher(&smConfig, nil, "")
	assert.Nil(err, "Can't create token watcher")

	crManager := service.CRManagerImpl{
		SMConfig:      &smConfig,
		GetHttpClient: &httpClientMock,
		TokenWatcher: tokenWatcher,
	}

	// Add services with CR cycles
	serviceE := "serviceE"
	serviceF := "serviceF"
	smConfig.Testing.SMDict.Services[serviceE] = model.SMObject{
		After:      []string{},
		Before:     []string{serviceF},
		Parameters: parameters,
	}
	smConfig.Testing.SMDict.Services[serviceF] = model.SMObject{
		After:      []string{serviceE},
		Before:     []string{},
		Parameters: parameters,
	}

	expectedStatusE := model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{},
			Before: []string{serviceF},
		},
	}

	expectedStatusF := model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{serviceE},
			Before: []string{},
		},
	}

	// Check that status procedure works with status
	serviceEStatus, err := crManager.GetServiceStatus(context.Background(), &serviceE, true)
	assert.Nil(err, "Can't get status for %s", serviceE)
	assert.Equal(&model.SMStatusResponse{
		Services: map[string]model.SMStatus{
			serviceE: expectedStatusE,
			serviceF: expectedStatusF,
		},
	}, serviceEStatus, "status for serviceE is not desired")
}

func TestCRManager_DepsCycles(t *testing.T) {
	_ = envconfig.InitConfig()
	assert := require.New(t)

	httpClientMock := mock.HttpClientMock{
		ServiceStatus: model.ServiceSiteManagerResponse{Mode: "active", Status: "done"},
		ServiceHealth: model.ServiceHealthzResponse{Status: "up"},
	}
	tokenWatcher, err := service.NewTokenWatcher(&smConfig, nil, "")
	assert.Nil(err, "Can't create token watcher")

	crManager := service.CRManagerImpl{
		SMConfig:      &smConfig,
		GetHttpClient: &httpClientMock,
		TokenWatcher: tokenWatcher,
	}

	// Add services with deps cycles
	serviceE := "serviceE"
	serviceF := "serviceF"
	smConfig.Testing.SMDict.Services[serviceE] = model.SMObject{
		After:      []string{serviceF},
		Before:     []string{serviceF},
		Parameters: parameters,
	}
	smConfig.Testing.SMDict.Services[serviceF] = model.SMObject{
		After:      []string{},
		Before:     []string{},
		Parameters: parameters,
	}

	expectedStatusE := model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{serviceF},
			Before: []string{serviceF},
		},
	}

	expectedStatusF := model.SMStatus{
		Mode:   "active",
		Status: "done",
		Health: "up",
		Deps: &model.SMStatusDeps{
			After:  []string{},
			Before: []string{},
		},
	}

	// Check that status procedure works with status
	serviceEStatus, err := crManager.GetServiceStatus(context.Background(), &serviceE, true)
	assert.Nil(err, "Can't get status for %s", serviceE)
	assert.Equal(&model.SMStatusResponse{
		Services: map[string]model.SMStatus{
			serviceE: expectedStatusE,
			serviceF: expectedStatusF,
		},
	}, serviceEStatus, "status for serviceE is not desired")
}
