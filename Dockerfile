FROM golang:1.20-alpine AS build-kubedoom
WORKDIR /go/src/kubedoom
COPY go.mod kubedoom.go ./
RUN apk add --no-cache git && \
  CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o kubedoom .

FROM ubuntu:22.04 AS build-essentials
ARG TARGETARCH
ARG KUBECTL_VERSION=1.24.12
RUN apt-get update && \
  apt-get install -y -o APT::Install-Suggests=0 --no-install-recommends \
  wget ca-certificates && \
  rm -rf /var/lib/apt/lists/* && \
  wget http://distro.ibiblio.org/pub/linux/distributions/slitaz/sources/packages/d/doom1.wad && \
  wget -O /usr/bin/kubectl "https://storage.googleapis.com/kubernetes-release/release/v${KUBECTL_VERSION}/bin/linux/${TARGETARCH}/kubectl" && \
  chmod +x /usr/bin/kubectl

FROM ubuntu:22.04 AS build-doom
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && \
  apt-get install -y -o APT::Install-Suggests=0 --no-install-recommends \
  build-essential \
  libsdl-mixer1.2-dev \
  libsdl-net1.2-dev \
  gcc && \
  rm -rf /var/lib/apt/lists/*
COPY /dockerdoom /dockerdoom
WORKDIR /dockerdoom/trunk
RUN ls -la
RUN ./configure && make && make install

FROM ubuntu:22.04 as build-converge
COPY --from=build-essentials /doom1.wad /root
COPY --from=build-essentials /usr/bin/kubectl /usr/bin
COPY --from=build-kubedoom /go/src/kubedoom/kubedoom /usr/bin
COPY --from=build-doom /usr/local/games/psdoom /usr/local/games

FROM ubuntu:22.04
ARG VNCPASSWORD=idbehold
RUN apt-get update && \
  apt-get install -y -o APT::Install-Suggests=0 --no-install-recommends \
  libsdl-mixer1.2 \
  libsdl-net1.2 \
  x11vnc \
  xvfb \
  netcat-openbsd && \
  rm -rf /var/lib/apt/lists/* && \
  mkdir /root/.vnc && x11vnc -storepasswd "${VNCPASSWORD}" /root/.vnc/passwd
COPY --from=build-converge /usr /usr
WORKDIR /root
ENTRYPOINT ["/usr/bin/kubedoom"]
