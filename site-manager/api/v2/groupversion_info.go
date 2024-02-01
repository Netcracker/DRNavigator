package v2

import (
	envconfig "github.com/netcracker/drnavigator/site-manager/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// AddToScheme adds the types in this group-version to the given scheme.
func AddToScheme(s *runtime.Scheme) error {
	groupVersion := schema.GroupVersion{Group: envconfig.EnvConfig.CRGroup, Version: CRVersion}
	s.AddKnownTypeWithName(groupVersion.WithKind(envconfig.EnvConfig.CRKind), &CR{})
	s.AddKnownTypeWithName(groupVersion.WithKind(envconfig.EnvConfig.CRKindList), &CRList{})
	metav1.AddToGroupVersion(s, groupVersion)
	return nil
}
