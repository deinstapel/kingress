package main

import (
  log "github.com/sirupsen/logrus"
  "os"
  "os/exec"
  "syscall"
)

const haproxyConfFile = "/haproxy.cfg"
const haproxySocket = "/haproxy.sock"
var haproxyBinary string
var haproxyProcess *exec.Cmd

func init() {
  bin, err := exec.LookPath("haproxy")
  if err != nil {
    log.WithField("err", err).Fatal("couldn't find haproxy")
  }
  haproxyBinary = bin

}

func startHaProxy() {
  cmdline := []string{"-db", "-f", haproxyConfFile, "-W"}
  cmd := exec.Command(haproxyBinary, cmdline...)
  cmd.Stderr = os.Stderr
  cmd.Stdout = os.Stdout
  haproxyProcess = cmd
  if err := cmd.Run(); err != nil {
    if (cmd.ProcessState != nil) {
      log.WithField("err", err).WithField("code", cmd.ProcessState.ExitCode()).Fatal("haproxy exited")
    } else {
      log.WithField("err", err).Fatal("HAProxy start failed")
    }
  }
}

func stopHaProxy() {
  if haproxyProcess != nil && haproxyProcess.Process != nil {
    haproxyProcess.Process.Signal(syscall.SIGTERM)
  }
}

func reloadHaProxy() {
  if haproxyProcess == nil {
    go startHaProxy()
    return
  }
  if err := haproxyProcess.Process.Signal(syscall.SIGUSR2); err != nil {
    log.WithField("err", err).Warning("failed to reload haproxy")
  } else {
    log.Info("reloaded haproxy")
  }
}
