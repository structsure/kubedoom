FORCE:

TAG:=localhost:5000/kubedoom:latest

devtools:
	brew install colima kubernetes-cli k9s yq helm

stop:
	colima stop -p kubedoom

cluster:
	colima start -p kubedoom --cpu 8 --memory 19 --kubernetes

registry:
	docker run -d -p 5000:5000 --restart=always --name registry registry

build:
	docker build -t $(TAG) .

flux:
	./bigbang/scripts/install_flux.sh -u `yq .registryCredentials.username ib_creds.yaml`  -p `yq .registryCredentials.password ib_creds.yaml`

bigbang: FORCE
	helm upgrade --install bigbang ./bigbang/chart \
  --values https://repo1.dso.mil/platform-one/big-bang/bigbang/-/raw/master/chart/ingress-certs.yaml \
  --values ./ib_creds.yaml \
  --values ./values.yaml \
  --namespace=bigbang --create-namespace

deploy: build
	docker push $(TAG)
	kubectl apply -k manifest/dev
	kubectl rollout restart -n kubedoom deployments/kubedoom
	kubectl rollout status -n kubedoom deployments/kubedoom
