## dingo tool 

dingo工具是 Dingo 团队为了提高系统的易用性，解决旧工具种类多输出繁琐等问题而设计的工具，
主要用于对Dingo文件存储集群进行运维的工具。

## Build dingo-tools

### Dependencies

#### Download dep

```sh
cd dingocli
git submodule sync
git submodule update --init --recursive
```
#### Install protobuf

```sh
PB_REL="https://github.com/protocolbuffers/protobuf/releases"
curl -LO $PB_REL/download/v25.1/protoc-25.1-linux-x86_64.zip
unzip protoc-25.1-linux-x86_64.zip -d $HOME/.local
export PATH="$PATH:$HOME/.local/bin"
```
#### Install gcc
##### Rocky 8.9/9.3
```sh
sudo dnf install -y epel-release
sudo dnf install -y gcc-toolset-13*
source /opt/rh/gcc-toolset-13/enable
```
##### Ubuntu 22.04/24.04
```sh
sudo apt update
sudo apt install -y  gcc g++
```

#### Install golang

```shell
wget https://go.dev/dl/go1.24.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.24.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Build 
```sh
cd dingocli

make build
```

### User Guide

[用户使用指南](./docs/userguide.md)