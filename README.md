## dingo tool 

dingo工具是 Dingo 团队为了提高系统的易用性，解决旧工具种类多输出繁琐等问题而设计的工具，
主要用于对Dingo文件存储集群进行运维的工具。

## Build dingo-tools

### Dependencies

#### Download dep

```sh
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

#### Install musl-gcc

```shell
wget https://musl.libc.org/releases/musl-1.2.5.tar.gz
tar -xzvf musl-1.2.5.tar.gz
cd musl-1.2.5 && sudo ./configure && sudo make install
export PATH=$PATH:/usr/local/musl/bin
```

#### Install golang

```shell
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Build 
```sh
cd dingofs-tools

make build
```

### User Guide 

如果元数据使用mdsv2版本的元数据,需要配置环境变量MDS_API_VERSION=2:
```sh
export MDS_API_VERSION=2
```

[用户使用指南](./docs/userguide.md)