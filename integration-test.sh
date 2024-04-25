#!/bin/bash

set -ve

# run the rclone integration tests against all the remotes

export GOPATH=~/go
export GO111MODULE=off

# checkout

mkdir -p ${GOPATH}/src/github.com/rclone/rclone

cd ${GOPATH}/src/github.com/rclone/rclone
# tidy up from previous runs
rm -f fs/operations/operations.test fs/sync/sync.test fs/test_all.log summary test.log
if [ -e ".git" ]; then
    git stash --include-untracked # stash any local changes just in case
    git checkout master
    git pull
else
    git clone https://github.com/rclone/rclone.git .
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
for i in 1 2 3 4 5; do GO111MODULE=on go get github.com/minio/minio@latest command && break || sleep 2; done
minio server --address 127.0.0.1:9000 ${minio_dir} >/tmp/minio.log 2>&1 &
minio_pid=$!

# make sure we build the optional extras

export GOTAGS=cmount

# run the tests

go install github.com/rclone/rclone/fstest/test_all

test_all -verbose -upload "pub.rclone.org:integration-tests" -email "nick@craig-wood.com" -output "/home/rclone/integration-test/rclone-integration-tests"

# stop minio
kill $minio_pid
