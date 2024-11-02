package test

import (
	"testing"

	crv1 "github.com/netcracker/drnavigator/site-manager/api/v1"
	crv2 "github.com/netcracker/drnavigator/site-manager/api/v2"
	crv3 "github.com/netcracker/drnavigator/site-manager/api/v3"
	"github.com/netcracker/drnavigator/site-manager/config"
	test_objects "github.com/netcracker/drnavigator/site-manager/tests/data"
	"github.com/stretchr/testify/require"
)

func TestConverter_ConvertV2ToV3_stateful(t *testing.T) {
	// Test conversion from v2 to v3
	_ = config.InitConfig()
	assert := require.New(t)

	// Check stateful service conversion
	testService := &crv2.CR{}
	test_objects.ServiceV2.DeepCopyInto(testService)
	testService.Spec.SiteManager.Module = "stateful"

	convertedService := &crv3.CR{}
	convertedService.TypeMeta = test_objects.ServiceV3.TypeMeta
	err := testService.ConvertTo(convertedService)
	assert.NoError(err, "conversion error")

	assert.Equal(test_objects.ServiceV3.TypeMeta, convertedService.TypeMeta, "type meta is not equal")
	assert.Equal(testService.ObjectMeta, convertedService.ObjectMeta, "object meta is not actual")

	assert.Nil(convertedService.Spec.SiteManager.Alias, "alias is desired")

	assert.Equal(testService.Spec.SiteManager.Module, convertedService.Spec.SiteManager.Module, "module is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.After, convertedService.Spec.SiteManager.After, "after is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Before, convertedService.Spec.SiteManager.Before, "before is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Sequence, convertedService.Spec.SiteManager.Sequence, "sequence is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.AllowedStandbyStateList, convertedService.Spec.SiteManager.AllowedStandbyStateList, "allowedStandbyStateList is not equal")
	assert.Equal(*testService.Spec.SiteManager.Timeout, *convertedService.Spec.SiteManager.Timeout, "timeout is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.ServiceEndpoint, convertedService.Spec.SiteManager.Parameters.ServiceEndpoint, "service endpoint is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.HealthzEndpoint, convertedService.Spec.SiteManager.Parameters.HealthzEndpoint, "health endpoint is not equal")

	assert.Equal(testService.Status.Summary, convertedService.Status.Summary, "summary status is not equal")
	assert.Equal(testService.Status.ServiceName, convertedService.Status.ServiceName, "service name in status is not equal")
}

func TestConverter_ConvertV2ToV3_not_stateful(t *testing.T) {
	// Test conversion from v2 to v3
	_ = config.InitConfig()
	assert := require.New(t)

	// Check stateful service conversion
	testService := &crv2.CR{}
	test_objects.ServiceV2.DeepCopyInto(testService)

	convertedService := &crv3.CR{}
	convertedService.TypeMeta = test_objects.ServiceV3.TypeMeta
	err := testService.ConvertTo(convertedService)
	assert.NoError(err, "conversion error")

	assert.Equal(test_objects.ServiceV3.TypeMeta, convertedService.TypeMeta, "type meta is not equal")
	assert.Equal(testService.ObjectMeta, convertedService.ObjectMeta, "object meta is not actual")

	assert.Equal(testService.Name, *convertedService.Spec.SiteManager.Alias, "alias is not desired")

	assert.Equal(testService.Spec.SiteManager.Module, convertedService.Spec.SiteManager.Module, "module is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.After, convertedService.Spec.SiteManager.After, "after is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Before, convertedService.Spec.SiteManager.Before, "before is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Sequence, convertedService.Spec.SiteManager.Sequence, "sequence is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.AllowedStandbyStateList, convertedService.Spec.SiteManager.AllowedStandbyStateList, "allowedStandbyStateList is not equal")
	assert.Equal(*testService.Spec.SiteManager.Timeout, *convertedService.Spec.SiteManager.Timeout, "timeout is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.ServiceEndpoint, convertedService.Spec.SiteManager.Parameters.ServiceEndpoint, "service endpoint is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.HealthzEndpoint, convertedService.Spec.SiteManager.Parameters.HealthzEndpoint, "health endpoint is not equal")

	assert.Equal(testService.Status.Summary, convertedService.Status.Summary, "summary status is not equal")
	assert.Equal(testService.Status.ServiceName, convertedService.Status.ServiceName, "service name in status is not equal")
}

func TestConverter_ConvertV1ToV3(t *testing.T) {
	// Test conversion from v1 to v3
	_ = config.InitConfig()
	assert := require.New(t)

	testService := &crv1.CR{}
	test_objects.ServiceV1.DeepCopyInto(testService)

	convertedService := &crv3.CR{}
	convertedService.TypeMeta = test_objects.ServiceV3.TypeMeta
	err := testService.ConvertTo(convertedService)
	assert.NoError(err, "conversion error")

	assert.Equal(test_objects.ServiceV3.TypeMeta, convertedService.TypeMeta, "type meta is not equal")
	assert.Equal(testService.ObjectMeta, convertedService.ObjectMeta, "object meta is not actual")

	assert.Nil(convertedService.Spec.SiteManager.Alias, "alias is desired")

	assert.Equal("stateful", convertedService.Spec.SiteManager.Module, "module is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.After, convertedService.Spec.SiteManager.After, "after is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Before, convertedService.Spec.SiteManager.Before, "before is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Sequence, convertedService.Spec.SiteManager.Sequence, "sequence is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.AllowedStandbyStateList, convertedService.Spec.SiteManager.AllowedStandbyStateList, "allowedStandbyStateList is not equal")
	assert.Equal(*testService.Spec.SiteManager.Timeout, *convertedService.Spec.SiteManager.Timeout, "timeout is not equal")
	assert.Equal(testService.Spec.SiteManager.ServiceEndpoint, convertedService.Spec.SiteManager.Parameters.ServiceEndpoint, "service endpoint is not equal")
	assert.Equal(testService.Spec.SiteManager.HealthzEndpoint, convertedService.Spec.SiteManager.Parameters.HealthzEndpoint, "health endpoint is not equal")

	assert.Equal(testService.Status.Summary, convertedService.Status.Summary, "summary status is not equal")
	assert.Equal(testService.Status.ServiceName, convertedService.Status.ServiceName, "service name in status is not equal")
}

func TestConverter_ConvertV3ToV2(t *testing.T) {
	// Test conversion from v2 to v1
	_ = config.InitConfig()
	assert := require.New(t)

	testService := &crv3.CR{}
	test_objects.ServiceV3.DeepCopyInto(testService)

	convertedService := &crv2.CR{}
	convertedService.TypeMeta = test_objects.ServiceV2.TypeMeta
	err := convertedService.ConvertFrom(testService)
	assert.NoError(err, "conversion error")

	assert.Equal(test_objects.ServiceV2.TypeMeta, convertedService.TypeMeta, "type meta is not equal")
	assert.Equal(testService.ObjectMeta, convertedService.ObjectMeta, "object meta is not actual")

	assert.Equal(testService.Spec.SiteManager.Module, convertedService.Spec.SiteManager.Module, "module is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.After, convertedService.Spec.SiteManager.After, "after is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Before, convertedService.Spec.SiteManager.Before, "before is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Sequence, convertedService.Spec.SiteManager.Sequence, "sequence is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.AllowedStandbyStateList, convertedService.Spec.SiteManager.AllowedStandbyStateList, "allowedStandbyStateList is not equal")
	assert.Equal(*testService.Spec.SiteManager.Timeout, *convertedService.Spec.SiteManager.Timeout, "timeout is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.ServiceEndpoint, convertedService.Spec.SiteManager.Parameters.ServiceEndpoint, "service endpoint is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.HealthzEndpoint, convertedService.Spec.SiteManager.Parameters.HealthzEndpoint, "health endpoint is not equal")
	assert.Equal("", convertedService.Spec.SiteManager.Parameters.IngressEndpoint, "ingress endpoint is not equal")

	assert.Equal(testService.Status.Summary, convertedService.Status.Summary, "summary status is not equal")
	assert.Equal(testService.Status.ServiceName, convertedService.Status.ServiceName, "service name in status is not equal")
}

func TestConverter_ConvertV3ToV1(t *testing.T) {
	// Test conversion from v3 to v1
	_ = config.InitConfig()
	assert := require.New(t)

	testService := &crv3.CR{}
	test_objects.ServiceV3.DeepCopyInto(testService)

	convertedService := &crv1.CR{}
	convertedService.TypeMeta = test_objects.ServiceV1.TypeMeta
	err := convertedService.ConvertFrom(testService)
	assert.NoError(err, "conversion error")

	assert.Equal(test_objects.ServiceV1.TypeMeta, convertedService.TypeMeta, "type meta is not equal")
	assert.Equal(testService.ObjectMeta, convertedService.ObjectMeta, "object meta is not actual")

	assert.ElementsMatch(testService.Spec.SiteManager.After, convertedService.Spec.SiteManager.After, "after is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Before, convertedService.Spec.SiteManager.Before, "before is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.Sequence, convertedService.Spec.SiteManager.Sequence, "sequence is not equal")
	assert.ElementsMatch(testService.Spec.SiteManager.AllowedStandbyStateList, convertedService.Spec.SiteManager.AllowedStandbyStateList, "allowedStandbyStateList is not equal")
	assert.Equal(*testService.Spec.SiteManager.Timeout, *convertedService.Spec.SiteManager.Timeout, "timeout is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.ServiceEndpoint, convertedService.Spec.SiteManager.ServiceEndpoint, "service endpoint is not equal")
	assert.Equal(testService.Spec.SiteManager.Parameters.HealthzEndpoint, convertedService.Spec.SiteManager.HealthzEndpoint, "health endpoint is not equal")
	assert.Equal("", convertedService.Spec.SiteManager.IngressEndpoint, "ingress endpoint is not equal")

	assert.Equal(testService.Status.Summary, convertedService.Status.Summary, "summary status is not equal")
	assert.Equal(testService.Status.ServiceName, convertedService.Status.ServiceName, "service name in status is not equal")
}
