package main

import (
  "context"
  "github.com/ericchiang/k8s"
  corev1 "github.com/ericchiang/k8s/apis/core/v1"
  "github.com/ghodss/yaml"
  log "github.com/sirupsen/logrus"
  "io/ioutil"
  "os"
  "os/signal"
  "strings"
)

func init() {
  log.SetFormatter(&log.TextFormatter{})
  log.SetOutput(os.Stderr)
  log.SetLevel(log.TraceLevel)
}

func makeKubeconfigClient(path string) (*k8s.Client, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := new(k8s.Config)
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}
	client, err := k8s.NewClient(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func makeClient() (*k8s.Client, error) {
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return makeKubeconfigClient(kubeconfig)
	}
	return k8s.NewInClusterClient()
}

func main() {
  client, err := makeClient()
  if err != nil {
    log.WithField("err", err).Fatal("Failed to create cluster client")
    os.Exit(1)
  }
  nodeName, ok := os.LookupEnv("NODE_NAME")
  if !ok {
    log.Fatal("NODE_NAME not set, please define it using downwards API")
    os.Exit(2)
  }

  ctx, cancel := context.WithCancel(context.Background())
  go watchNode(ctx, client, nodeName)
  go watchCrd(ctx, client)

  sigchan := make(chan os.Signal, 1)
  signal.Notify(sigchan, os.Interrupt)
  <-sigchan

  log.Info("Terminating")
  stopHaProxy()
  cancel()
}


func watchNode(ctx context.Context, client *k8s.Client, nodeName string) {
  for {
    log.Trace("start node watch")
    node := &corev1.Node{}
    nodeWatcher, err := client.Watch(ctx, k8s.AllNamespaces, node)
    if err != nil {
      log.WithField("err", err).Fatal("node watch failed, rbac ok?")
      break
    }

    for {
      innerNode := &corev1.Node{}
      t, err := nodeWatcher.Next(innerNode)
      if err != nil {
        if !strings.Contains(err.Error(), "EOF") {
          log.WithField("err", err).Fatal("node watch errored")
        } else {
          log.Debug("node watch ended, restarting")
          break
        }
      }
      if (t == k8s.EventDeleted) {
        continue
      }
      if (*innerNode.Metadata.Name != nodeName) {
        continue
      }
      regenerateListeners(true, innerNode.Metadata.Labels)
      log.WithField("labels", myLabels).Trace("label update")
    }
  }
}

func watchCrd(ctx context.Context, client *k8s.Client) {
  for {
    log.Trace("start crd watch")
    crdList := &ExposeConfig{}
    crdWatcher, err := client.Watch(ctx, k8s.AllNamespaces, crdList)
    if err != nil {
      log.WithField("err", err).Fatal("Failed to watch CRDs")
      break
    }

    for {
      exposeConfig := &ExposeConfig{}
      t, err := crdWatcher.Next(exposeConfig)
      if err != nil {
        if !strings.Contains(err.Error(), "EOF") {
          log.WithField("err", err).Fatal("node watch errored")
        } else {
          log.Debug("node watch ended, restarting")
          break
        }
      }
      if (t != k8s.EventDeleted) {
        upsertListener(exposeConfig)
      } else {
        deleteListener(exposeConfig)
      }
    }
  }
}
