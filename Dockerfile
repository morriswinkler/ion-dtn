FROM ubuntu:16.04

ARG DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
git gcc make wget mercurial automake autoconf patch nmap tcl libc6-dev nano tshark tcl-dev libdb-dev libssl-dev g++ \
bash-completion ca-certificates
RUN apt-get clean

# ion-dtn
WORKDIR /usr/local/src/ion
COPY . .

RUN ./configure --prefix=/usr/local \
    && make \
    && make install \
    && echo /usr/local/lib > /etc/ld.so.conf.d/local.conf \
    && ldconfig

# golang
WORKDIR /usr/local
RUN wget --no-check-certificate https://dl.google.com/go/go1.11.1.linux-amd64.tar.gz \
    && tar -xvzf go1.11.1.linux-amd64.tar.gz \
    && echo "export PATH=$PATH:/usr/local/go/bin:~/go/bin" >> /etc/bash.bashrc

# config
WORKDIR /ion
RUN cp -fr /usr/local/src/ion/configs/2node-stcp config
