package main

import (
  "k8s.io/apimachinery/pkg/labels"
  "k8s.io/apimachinery/pkg/selection"
)

func simpleToK8sLabels(lbl map[string]string) (labels.Selector, error) {
  if len(lbl) == 0 {
    return labels.Everything(), nil
  }
  lblOut := labels.NewSelector()
  for k, v := range lbl {
    req, err := labels.NewRequirement(k, selection.Equals, []string{v})
    if err != nil {
      return nil, err
    }
    lblOut = lblOut.Add(*req)
  }
  return lblOut, nil
}
