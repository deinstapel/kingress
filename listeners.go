package main

import (
	"k8s.io/apimachinery/pkg/labels"
  log "github.com/sirupsen/logrus"
  "text/template"
  "os"
)

const haproxyConfig = `
global
    maxconn     20000
    log         stdout format raw local0 info
    stats socket /haproxy.sock mode 600 level admin
    stats timeout 2m
    pidfile     /haproxy.pid

{{ range $name, $config := . }}
{{ if $config.Enabled }}

frontend {{ $name }}
    bind :{{ $config.ListenPort }}
    mode                 {{ $config.ListenProto}}
    log                  global
    option               {{ $config.ListenProto }}log
    option               dontlognull
    timeout              client  30s
    default_backend {{ $name }}

backend {{ $name }}
    mode        {{ $config.ListenProto }}
    balance     roundrobin
    timeout     connect 5s
    timeout     server  30s
    server      static {{ $config.Destination.ServiceName }}.{{ $config.Destination.Namespace }}:{{ $config.Destination.Port }} {{ if $config.EnableProxyProtocol }} send-proxy {{ end }}

{{ end }}
{{ end }}
`

var myLabels = map[string]string{}
var exposes = map[string]*ExposeConfig{}

var gotNode = false
func regenerateListeners(nodeMode bool, localLabels map[string]string) {
  localLog := log.WithFields(log.Fields{
    "module": "listener",
    "method": "regenerate",
  })
  localLog.Trace("start")
  if (nodeMode) {
    localLog.WithField("labels", localLabels).Trace("update labels")
    gotNode = true
    myLabels = localLabels
  }
  if !gotNode {
    localLog.Trace("no labels, skipping")
    return
  }

  for name, ep := range exposes {
    lbl, err := simpleToK8sLabels(ep.ExposeOn)
    if err != nil {
      localLog.WithFields(log.Fields{
        "err": err,
        "config": name,
      }).Warning("failed to build label selector, disabling")
      ep.Enabled = false
      continue
    }
    ep.Enabled = lbl.Matches(labels.Set(myLabels))
    localLog.WithField("config", name).WithField("enabled", ep.Enabled).Trace("parse ok")
  }

  proto := template.Must(template.New("haproxy-config").Parse(haproxyConfig))
  if err := proto.Execute(os.Stdout, exposes); err != nil {
    localLog.WithField("err", err).Warn("template rendering failed")
  }

  localLog.Trace("finish")
}

func upsertListener(ep *ExposeConfig) {
  localLog := log.WithFields(log.Fields{
    "module": "listener",
    "method": "upsert",
    "ep": *ep.Metadata.Name,
    "port": ep.ListenPort,
    "proto": ep.ListenProto,
    "destination": ep.Destination.ServiceName + "." + ep.Destination.Namespace,
    "proxy": ep.EnableProxyProtocol,
  })
  localLog.Debug("start")
  exposes[*ep.Metadata.Name] = ep
  regenerateListeners(false, nil)
  localLog.Trace("finish")
}
func deleteListener(ep *ExposeConfig) {
  localLog := log.WithFields(log.Fields{
    "module": "listener",
    "method": "delete",
    "ep": *ep.Metadata.Name,
  })
  localLog.Debug("start")
  delete(exposes, *ep.Metadata.Name)
  regenerateListeners(false, nil)
  localLog.Trace("finish")
}

