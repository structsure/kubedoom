#!sh

NS=$1
shift

kubectl get -n kube-system ks $NS -o json | jq '.spec.finalizers = []' | kubectl -n kube-system replace --raw "/api/v1beta2/kustomize.toolkit.fluxcd.io/$NS/finalize" -f -
