BUILD_DIR=$GOPATH/src/gitlab.alibaba-inc.com/kubedlpro/kubedlpro/APP-META/docker-config/

mkdir -p $BUILD_DIR
cd $GOPATH/src/gitlab.alibaba-inc.com/kubedlpro/kubedlpro || exit
go build -mod=vendor -o bin/manager ./
cp bin/manager $BUILD_DIR/environment/bin/
cp -r config/crd/* $BUILD_DIR/environment/cfg/
