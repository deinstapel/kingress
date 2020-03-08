package main

import (
	"github.com/ericchiang/k8s"
  metav1 "github.com/ericchiang/k8s/apis/meta/v1"
)

type ExposeConfig struct {
  Metadata *metav1.ObjectMeta `json:"metadata"`
  ListenPort int           `json:"port"`
  ListenProto string          `json:"proto"`
  Destination *ServiceSelector                   `json:"destination"`
  EnableProxyProtocol bool `json:"enableProxy"`
  ExposeOn *metav1.LabelSelector `json:"exposeOn"`
}

type ServiceSelector struct {
  ServiceName string `json:"service"`
  Namespace string `json:"namespace"`
  Port int `json:"port"`
}

func(e *ExposeConfig) GetMetadata() *metav1.ObjectMeta {
  return e.Metadata
}

type ExposeConfigList struct {
  Metadata *metav1.ListMeta `json:"metadata"`
  Items []ExposeConfig `json:"items"`
}

func(e *ExposeConfigList) GetMetadata() *metav1.ListMeta {
  return e.Metadata
}

func init() {
  k8s.Register("network.deinstapel.de", "v1alpha1", "exposeconfigs", false, &ExposeConfig{})
  k8s.RegisterList("network.deinstapel.de", "v1alpha1", "exposeconfigs", false, &ExposeConfigList{})
}
