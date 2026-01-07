# Ensure GOPATH/bin is in the PATH
export PATH=$PATH:$(go env GOPATH)/bin

# dingofs proto path
export PROTO_PATH="./third-party/dingofs-proto"

# proto dir
mkdir -p proto

## dingofs

### dingofs-proto/error.proto
protoc --experimental_allow_proto3_optional --go_out=proto --proto_path=${PROTO_PATH} \
    ${PROTO_PATH}/dingofs/error.proto

### dingofs-proto/mds.proto
protoc --experimental_allow_proto3_optional --go_out=proto --proto_path=${PROTO_PATH} \
    --go_opt=Mdingofs/error.proto=github.com/dingodb/dingofs-tools/proto/dingofs/proto/error \
    ${PROTO_PATH}/dingofs/mds.proto

### dingofs-proto/cachegroup.proto
protoc --experimental_allow_proto3_optional --go_out=proto --proto_path=${PROTO_PATH} \
    ${PROTO_PATH}/dingofs/cachegroup.proto

# grpc
## fs
protoc --experimental_allow_proto3_optional --go-grpc_out=proto --proto_path=${PROTO_PATH} ${PROTO_PATH}/dingofs/*.proto
