package main

import (
	"github.com/ericchiang/k8s"
  metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

// ExposeConfig is a struct defining a CRD which is used to
// expose a certain service at the edge of a cluster
type ExposeConfig struct {
  Metadata *metav1.ObjectMeta `json:"metadata"`
  ListenPort int           `json:"port"`
  ListenProto string          `json:"proto"`
  Destination *ServiceSelector                   `json:"destination"`
  EnableProxyProtocol bool `json:"enableProxy"`
  ExposeOn map[string]string `json:"exposeOn"`
  Enabled bool `json:"-"`
}

// ServiceSelector is a reference towards a service within a
// cluster in an arbitrary namespace.
type ServiceSelector struct {
  ServiceName string `json:"service"`
  Namespace string `json:"namespace"`
  Port int `json:"port"`
}

// GetMetadata gets the object metadata.
func(e *ExposeConfig) GetMetadata() *metav1.ObjectMeta {
  return e.Metadata
}

// ExposeConfigList is the corresponding list struct to
// ExposeConfig
type ExposeConfigList struct {
  Metadata *metav1.ListMeta `json:"metadata"`
  Items []ExposeConfig `json:"items"`
}


// GetMetadata gets the object metadata.
func(e *ExposeConfigList) GetMetadata() *metav1.ListMeta {
  return e.Metadata
}

func init() {
  k8s.Register("network.deinstapel.de", "v1alpha1", "exposeconfigs", false, &ExposeConfig{})
  k8s.RegisterList("network.deinstapel.de", "v1alpha1", "exposeconfigs", false, &ExposeConfigList{})
}
