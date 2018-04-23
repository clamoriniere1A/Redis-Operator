#!/bin/bash

GIT_ROOT=$(git rev-parse --show-toplevel)
. ${GIT_ROOT}/tools/demo-utils.sh

desc "Remove redis-operator helm chart"
run "helm del --purge operator"

desc "Remove namespaces"
run "kubectl delete ns operators test-ns"

desc "Remove service-catalog helm chart"
run "helm del --purge catalog"
run "helm repo remove svc-cat"