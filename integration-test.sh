#!/bin/bash

set -e

# run the rclone integration tests against all the remotes

export GOPATH=~/go

# checkout

mkdir -p ${GOPATH}/src/github.com/ncw/rclone

cd ${GOPATH}/src/github.com/ncw/rclone
# tidy up from previous runs
rm -f fs/operations/operations.test fs/sync/sync.test fs/test_all.log summary test.log
if [ -e ".git" ]; then
    git stash --include-untracked # stash any local changes just in case
    git checkout master
    git pull
else
    git clone https://github.com/ncw/rclone.git .
fi

# build rclone

make

# make sure restic is up to date for the cmd/serve/restic integration tests

go get -u github.com/restic/restic/...

# update and start minio

minio_dir=/tmp/minio

killall -q minio || true
rm -rf ${minio_dir}
mkdir -p ${minio_dir}
export MINIO_ACCESS_KEY=minio
export MINIO_SECRET_KEY=AxedBodedGinger7
go get -u github.com/minio/minio
minio server --address 127.0.0.1:9000 ${minio_dir} >/tmp/minio.log 2>&1 &
minio_pid=$!

# make sure we build the optional extras

export GOTAGS=cmount

# run the tests

go install github.com/ncw/rclone/fstest/test_all

test_all -upload "memstore:pub-rclone-org//integration-tests" -email "nick@craig-wood.com" -output "/home/rclone/integration-test/rclone-integration-tests"

# stop minio
kill $minio_pid
