#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail
set -x

export ROOT=$(dirname "${BASH_SOURCE}")/..
echo "TEST $CLUSTER"

printenv

ctl=/usr/local/bin/kubectl
if [ "$CLUSTER" == "openshift" ]; then
    echo "INIT Openshift test platform"
    ./hack/ci-openshift-install.sh
    ctl=/usr/local/bin/oc
else
    echo "INIT Kubernetes test platform"
    ./hack/ci-minikube-install.sh
fi

# "ctl" command is in fact "kubectl" or "oc" depending of the CLUSTER var env value
# common part
JSONPATH='{range .items[*]}{@.metadata.name}:{range @.status.conditions[*]}{@.type}={@.status};{end}{end}'; until $ctl get nodes -o jsonpath="$JSONPATH" 2>&1 | grep -q "Ready=True"; do sleep 1; done
$ctl get nodes
$ctl get pods --all-namespaces
curl https://raw.githubusercontent.com/kubernetes/helm/master/scripts/get | bash
$ctl -n kube-system create sa tiller
$ctl create clusterrolebinding tiller --clusterrole cluster-admin --serviceaccount=kube-system:tiller
helm init --wait --service-account tiller

# Install Service Catalog
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
until helm install --wait svc-cat/catalog --version 0.1.9 --name catalog --namespace catalog; do sleep 1; printf "."; done; echo
until [ $($ctl get deployment catalog-catalog-apiserver -n catalog -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
until [ $($ctl get deployment catalog-catalog-controller-manager -n catalog -ojsonpath="{.status.conditions[?(@.type=='Available')].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done; echo

make -C $ROOT build
make -C $ROOT test
make -C $ROOT TAG=$TAG container
docker tag redisoperator/redisnode:$TAG redisoperator/redisnode:4.0
docker tag redisoperator/redisnode:$TAG redisoperator/redisnode:latest
docker images

helm install --wait --version $TAG -n end2end-test --set broker.activate=true --set image.pullPolicy=IfNotPresent --set image.tag=$TAG chart/redis-operator
printf "Waiting for redis-operator deployment to complete."
until [ $($ctl get deployment end2end-test-redis-operator -ojsonpath="{.status.conditions[?(@.type==\"Available\")].status}") == "True" ] > /dev/null 2>&1; do sleep 1; printf "."; done
echo

$ctl logs -f $($ctl get pod --selector=app=end2end-test-redis-operator -n default --output=jsonpath={.items[0].metadata.name}) > /tmp/tmp.operator.logs &

$ctl get pods --all-namespaces

cd ./test/e2e
EXIT_CODE=0
go test -c && ./e2e.test --kubeconfig=$HOME/.kube/config --image-tag=$TAG --ginkgo.slowSpecThreshold 350 || EXIT_CODE=$? && true ;
helm delete end2end-test
cat /tmp/tmp.operator.logs
exit $EXIT_CODE
