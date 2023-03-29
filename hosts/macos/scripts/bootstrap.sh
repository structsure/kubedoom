#!zsh
export GITHUB_TOKEN=$(security find-internet-password -s github.com -w)
flux bootstrap github --owner=seagard --branch=flux --repository=kubedoom --path=clusters/bootstrap --personal --recurse-submodules
kubectl apply -f clusters/minikube/bigbang