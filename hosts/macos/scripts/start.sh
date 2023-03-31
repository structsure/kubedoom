#!zsh
echo not working
exit 1
# maybe host only cidr, despite wrong driver
# kube system is commint up on a diffrernt cidr than the rest of the services
# --host-only-cidr string             The CIDR to be used for the minikube VM (virtualbox driver only) (default "192.168.59.1/24")
#  maybe set --extra-args kubelet.pod-cidr=<string> 
# found CIDR 10.96.0.0/12
minikube start \ 
--driver=hyperkit \ 
--cpus 8 \ 
--memory 16384 \ 
--embed-certs \ 
# --extra-config=kubelet.ClusterCIDR=192.168.0.0/16 \
# --extra-config=proxy.ClusterCIDR=192.168.0.0/16 \
# --extra-config=controller-manager.ClusterCIDR=192.168.0.0/16 

