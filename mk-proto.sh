# Ensure GOPATH/bin is in the PATH
export PATH=$PATH:$(go env GOPATH)/bin

# dingofs proto path
export PROTO_PATH="./third-party/dingofs-proto"

# proto dir
mkdir -p proto

## dingofs
### dingofs-proto/cli2.proto
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/common.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/common \
    ${PROTO_PATH}/dingofs/cli2.proto

### dingofs-proto/common.proto
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    ${PROTO_PATH}/dingofs/common.proto

### dingofs-proto/copyset.proto
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/common.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/common \
    ${PROTO_PATH}/dingofs/copyset.proto

### dingofs-proto/heartbeat.proto
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/common.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/common \
    --go_opt=Mdingofs/proto/heartbeat.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/heartbeat \
    ${PROTO_PATH}/dingofs/heartbeat.proto

### dingofs-proto/mds.proto
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/common.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/common \
    --go_opt=Mdingofs/topology.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/topology \
    ${PROTO_PATH}/dingofs/mds.proto

### dingofs-proto/metaserver.proto 
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/common.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/common \
    ${PROTO_PATH}/dingofs/metaserver.proto

### dingofs-proto/schedule.proto
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    ${PROTO_PATH}/dingofs/schedule.proto

### dingofs-proto/topology.proto
protoc --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/common.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/common \
    --go_opt=Mdingofs/heartbeat.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/heartbeat \
    ${PROTO_PATH}/dingofs/topology.proto

### dingofs-proto/error.proto
protoc --experimental_allow_proto3_optional --go_out=proto --proto_path=${PROTO_PATH} \
    ${PROTO_PATH}/dingofs/error.proto

### dingofs-proto/mdsv2.proto
protoc --experimental_allow_proto3_optional --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/error.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/error \
    ${PROTO_PATH}/dingofs/mdsv2.proto

### dingofs-proto/cachegroup.proto
protoc --experimental_allow_proto3_optional --go_out=proto --proto_path=${PROTO_PATH} \
    ${PROTO_PATH}/dingofs/cachegroup.proto

# grpc
## fs
protoc --experimental_allow_proto3_optional --go-grpc_out=proto --proto_path=${PROTO_PATH} ${PROTO_PATH}/dingofs/*.proto