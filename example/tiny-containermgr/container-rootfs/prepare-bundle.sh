#!/bin/bash

mkdir rootfs
sudo bash -c 'docker export $(docker create busybox) | tar -C rootfs -xvf -'
runc spec
sed -i 's/"sh"/"sh", "entrypoint.sh"/' config.json
sed -i 's/"terminal": true/"terminal": false/' config.json
sed -i 's/"readonly": true/"readonly": false/' config.json
