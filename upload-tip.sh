#!/bin/bash

set -e

# upload the tip website

export GOPATH=~/go
cd ${GOPATH}/src/github.com/rclone/rclone
make commanddocs
make website
rclone sync docs/public tip.rclone.org:
git checkout -- docs/content
echo "Uploaded new website to https://tip.rclone.org/"
