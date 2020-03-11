package main

import (
  log "github.com/sirupsen/logrus"
  "os/exec"
  "os"
  "strconv"
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
func restartHaProxy() {
  cmdline := []string{"-db", "-f", haproxyConfFile}
  if haproxyProcess != nil && haproxyProcess.ProcessState == nil && haproxyProcess.Process != nil {
    cmdline = append(cmdline, "-sf", strconv.Itoa(haproxyProcess.Process.Pid))
  }
  cmd := exec.Command(haproxyBinary, cmdline...)
  cmd.Stderr = os.Stderr
  cmd.Stdout = os.Stdout
  if err := cmd.Start(); err != nil {
    log.WithField("err", err).Fatal("HAProxy start failed")
  }
  go func() {
    if err := cmd.Wait(); err != nil {
      log.WithField("err", err).WithField("code", cmd.ProcessState.ExitCode()).Fatal("HAProxy error")
    }
    log.Info("HAProxy ended with code 0.")
  }()
  haproxyProcess = cmd
}
