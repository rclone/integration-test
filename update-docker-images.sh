#!/bin/bash

for image in $(docker images |grep -v REPOSITORY|awk '{print $1}'); do
    echo "Updating $image"
    docker pull $image
done
