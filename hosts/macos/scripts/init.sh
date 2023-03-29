#!zsh
brew install minikube kubectl hyperkit helm yq fluxcd/tap/flux ripgrep k9s

yq -n ".registryCredentials.username = $(security find-generic-password -s registry1.dso.mil |rg '.*acc.*(".*")$' -r '$1') \
      |.registryCredentials.password = \"$(security find-generic-password -s registry1.dso.mil -w)\"" \
> clusters/minikube/bigbang-values/secrets/ib_creds.yaml

echo -----
echo cheking livelness of your harbor api creds
echo -----
helm registry login registry1.dso.mil \
-u $(security find-generic-password -s registry1.dso.mil |rg '.*acc.*"(.*)"$' -r '$1') \
-p $(security find-generic-password -s registry1.dso.mil -w)

minikube start --driver=hyperkit
