package main

import (
  log "github.com/sirupsen/logrus"
  "k8s.io/apimachinery/pkg/labels"
  "net"
  "os"
  "strconv"
  "text/template"
  "time"
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
    timeout              client  1800s
    default_backend {{ $name }}

backend {{ $name }}
    mode        {{ $config.ListenProto }}
    balance     roundrobin
    timeout     connect 5s
    timeout     server  1800s
    server      static {{ $config.Destination.ServiceName }}.{{ $config.Destination.Namespace }}:{{ $config.Destination.Port }} {{ if $config.EnableProxyProtocol }} send-proxy {{ end }}

{{ end }}
{{ end }}
`

var myLabels = map[string]string{}
var exposes = map[string]*ExposeConfig{}

func checkReachability(ep *ExposeConfig) bool {
  // From https://stackoverflow.com/a/56336811

  timeout := time.Second
  host := ep.Destination.ServiceName + "." + ep.Destination.Namespace
  port := strconv.Itoa(ep.Destination.Port)

  conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port), timeout)
  if err != nil {
    log.WithFields(log.Fields{
      "err": err,
      "config": ep.Metadata.Name,
    }).Warning("Failed to reach destination service")

    return false
  }

  defer conn.Close()
  return true
}

func checkEnabled(ep *ExposeConfig) bool {
  lbl, err := simpleToK8sLabels(ep.ExposeOn)
  if err != nil {
    log.WithFields(log.Fields{
      "err": err,
      "config": ep.Metadata.Name,
    }).Warning("Failed to build label selector, disabling")
    return false
  }

  return lbl.Matches(labels.Set(myLabels))
}

var gotNode = false
func regenerateListeners(nodeMode bool, localLabels map[string]string) {
  localLog := log.WithFields(log.Fields{
    "module": "listener",
    "method": "regenerate",
  })
  localLog.Trace("start")
  defer localLog.Trace("finish")

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
    isReachable := checkReachability(ep)
    isEnabled := checkEnabled(ep)
    ep.Enabled = isEnabled && isReachable
    localLog.WithField("config", name).WithField("enabled", ep.Enabled).Trace("parse ok")
  }

  proto := template.Must(template.New("haproxy-config").Parse(haproxyConfig))
  cfgFile, err := os.Create(haproxyConfFile)
  if err != nil {
    localLog.WithField("err", err).Warn("open cfg file failed")
    return
  }
  if err := proto.Execute(cfgFile, exposes); err != nil {
    localLog.WithField("err", err).Warn("template rendering failed")
  }
  cfgFile.Close()
  reloadHaProxy()
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
  localLog.Trace(len(exposes))
  regenerateListeners(false, nil)
  localLog.Trace("finish")
}

