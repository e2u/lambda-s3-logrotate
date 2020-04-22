CURRENT_TIME=$(shell date "+%Y%m%d%H%M%S")
PWD=$(shell pwd)
BUILD_DIR=$(PWD)/build
SRC_DIR=$(PWD)/lambda
NOW=$(shell date "+%Y%m%d%H%M%S")
BUILD_TIME=$(shell date "+%Y-%m-%dT%H:%M:%S%z")
GIT_COMMIT_ID=$(shell git rev-parse --short HEAD)
LDFLAGS="-X main.GitCommitId=${GIT_COMMIT_ID} -X main.BuildTime=${BUILD_TIME}"

OBJ_NAME=s3-logrotate
LAMBDA_FUNC_NAME=s3-logrotate


.PHONY: default
default: help


# help 提取注释作为帮助信息
help:                              ## Show this help.
	@fgrep -h "##" $(MAKEFILE_LIST) | fgrep -v fgrep | sed -e 's/\\$$//' | sed -e 's/##//'

## clean 删除构建目录
.PHONY: clean
clean:
	rm -rf ${BUILD_DIR}


## run-srv 运行开发环境服務端
.PHONY: run-srv
run-srv:
	cd ${SRC_DIR} && _LAMBDA_SERVER_PORT=16000 go run main.go

## run-cli 運行本地測試客戶端
.PHONY: run-cli
run-cli:
	cd cli && go run main.go

## build 构建当前系统环境二进制文件
.PHONY: build
build: clean
	GOOS="linux" GOARCH="amd64"  go build -ldflags ${LDFLAGS} -o ${BUILD_DIR}/${OBJ_NAME} ${SRC_DIR}/*.go

## upx 压缩二进制文件
.PHONY: upx
upx:
	upx  ${BUILD_DIR}/${OBJ_NAME}


## deploy 部署到 lambda
.PHONY: deploy
deploy: build upx
	cd ${BUILD_DIR} && zip ${OBJ_NAME}.zip ${OBJ_NAME} && \
	aws --output=table lambda  update-function-code --function-name="${LAMBDA_FUNC_NAME}" --zip-file="fileb://${BUILD_DIR}/${OBJ_NAME}.zip"