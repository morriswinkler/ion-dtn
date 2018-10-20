FROM ubuntu:16.04

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
git gcc make wget mercurial automake autoconf patch nmap tcl libc6-dev nano tshark tcl-dev libdb-dev libssl-dev g++
RUN apt-get clean

WORKDIR /usr/local/src/ion
COPY . .

RUN ./configure --prefix=/usr/local \
    && make \
    && make install \
    && echo /usr/local/lib > /etc/ld.so.conf.d/local.conf \
    && ldconfig

WORKDIR /ion
RUN cp -fr /usr/local/src/ion/configs/2node-stcp config
