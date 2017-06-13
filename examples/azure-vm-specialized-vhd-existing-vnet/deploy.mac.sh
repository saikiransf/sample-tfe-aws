#!/bin/bash

set -o errexit -o nounset

if docker -v; then

  # generate a unique string for CI deployment
  export KEY=$(cat /dev/urandom | env LC_CTYPE=C tr -cd 'a-z' | head -c 12)
  export PASSWORD=$KEY$(cat /dev/urandom | env LC_CTYPE=C tr -cd 'A-Z' | head -c 2)$(cat /dev/urandom | env LC_CTYPE=C tr -cd '0-9' | head -c 2)
  export EXISTING_RESOURCE_GROUP=permanent
  export EXISTING_LINUX_IMAGE_URI=https://tfpermstor.blob.core.windows.net/vhds/osdisk_fmF5O5MxlR.vhd
  export EXISTING_STORAGE_ACCOUNT_NAME=tfpermstor
  export EXISTING_VIRTUAL_NETWORK_NAME=permanent-vnet
  export EXISTING_SUBNET_NAME=permanent-subnet
  export EXISTING_SUBNET_ID=/subscriptions/0baf16e2-035c-41fa-aadd-7259dccda244/resourceGroups/permanent/providers/Microsoft.Network/virtualNetworks/permanent-vnet/subnets/permanent-subnet

  /bin/sh ./deploy.ci.sh

else
  echo "Docker is used to run terraform commands, please install before run:  https://docs.docker.com/docker-for-mac/install/"
fi