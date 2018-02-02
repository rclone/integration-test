#!/bin/bash

set -e

# run the rclone integration tests against all the remotes

export GOPATH=~/go

# checkout

mkdir -p ${GOPATH}/src/github.com/ncw/rclone

cd ${GOPATH}/src/github.com/ncw/rclone
if [ -e ".git" ]; then
    git checkout master
    git pull
else
    git clone https://github.com/ncw/rclone.git .
fi

# run the tests

make test || true

# set TAG
make vars | grep TAG >sourceme
source sourceme
rm sourceme

# make paths
outstore=memstore:pub-rclone-org
outroot=integration-tests
outpath=${outroot}/`date +'%Y-%m/%Y-%m-%d-%H%M%S'`
www=https://pub.rclone.org/${outpath}
out=${outstore}/${outpath}
outtop=${outstore}/${outroot}

# make summary
rm -f summary
touch summary
echo >> summary "--------------------------------------------------------------" 
echo >> summary "go test results - full results at ${www}/test-${TAG}.txt"
echo >> summary "--------------------------------------------------------------" 
grep >> summary 'FAIL\|WARN' test.log || echo >> summary "No test failures"
echo >> summary
echo >> summary "--------------------------------------------------------------" 
echo >> summary "fs/test_all results - full results at ${www}/test_all-${TAG}.txt"
echo >> summary "--------------------------------------------------------------" 
grep >> summary -A1000 SUMMARY fs/test_all.log || echo >> summary "No SUMMARY found"

mail nick@craig-wood.com -s "rclone integration tests" <summary

# copy the results
rclone copyto test.log ${out}/test-${TAG}.txt
rclone copyto fs/test_all.log ${out}/test_all-${TAG}.txt
rclone copyto summary ${out}/summary-${TAG}.txt

# server side copy these to current
rclone copyto ${out}/test-${TAG}.txt ${outtop}/current-test.txt
rclone copyto ${out}/test_all-${TAG}.txt ${outtop}/current-test_all.txt
rclone copyto ${out}/summary-${TAG}.txt ${outtop}/current-summary.txt
