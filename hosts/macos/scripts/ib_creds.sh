#!zsh

yq -n ".registryCredentials.username = $(security find-generic-password -s registry1.dso.mil |rg '.*acc.*(".*")$' -r '$1') \
      |.registryCredentials.password = \"$(security find-generic-password -s registry1.dso.mil -w)\"" \
> ib_creds.yaml
