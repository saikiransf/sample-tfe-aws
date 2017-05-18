#!/bin/bash

set -o errexit -o nounset

docker run --rm -it \
  -e ARM_CLIENT_ID \
  -e ARM_CLIENT_SECRET \
  -e ARM_SUBSCRIPTION_ID \
  -e ARM_TENANT_ID \
  -e AAD_CLIENT_ID \
  -e AAD_CLIENT_SECRET \
  -e KEY_ENCRYPTION_KEY_URL \
  -e KEY_VAULT_RESOURCE_ID \
  -v $(pwd):/data \
  --entrypoint "/bin/sh" \
  hashicorp/terraform:light \
  -c "cd /data; \
      /bin/terraform get; \
      /bin/terraform validate; \
      /bin/terraform plan -out=out.tfplan \
        -var resource_group=$KEY \
        -var hostname=$KEY \
        -var admin_username=$KEY \
        -var admin_password=$PASSWORD \
        -var passphrase=$PASSWORD \
        -var key_vault_name=$KEY_VAULT_NAME \
        -var key_encryption_key_url=$KEY_ENCRYPTION_KEY_URL \
        -var key_vault_resource_id=$KEY_VAULT_RESOURCE_ID > /dev/null; \
      /bin/terraform apply out.tfplan > /dev/null"

# cleanup deployed azure resources via azure-cli
docker run --rm -it \
  azuresdk/azure-cli-python \
  sh -c "az login --service-principal -u $ARM_CLIENT_ID -p $ARM_CLIENT_SECRET --tenant $ARM_TENANT_ID > /dev/null; \
         az vm show -g $KEY -n $KEY; \
         az vm encryption show -g $KEY -n $KEY;"

# cleanup deployed azure resources via terraform
docker run --rm -it \
  -e ARM_CLIENT_ID \
  -e ARM_CLIENT_SECRET \
  -e ARM_SUBSCRIPTION_ID \
  -e ARM_TENANT_ID \
  -v $(pwd):/data \
  --workdir=/data \
  --entrypoint "/bin/sh" \
  hashicorp/terraform:light \
  -c "/bin/terraform destroy -force \
    -var resource_group=$KEY \
    -var hostname=$KEY \
    -var admin_username=$KEY \
    -var admin_password=$PASSWORD \
    -var passphrase=$PASSWORD \
    -var key_vault_name=$KEY_VAULT_NAME \
    -var key_encryption_key_url=$KEY_ENCRYPTION_KEY_URL \
    -var key_vault_resource_id=$KEY_VAULT_RESOURCE_ID \
    -target=azurerm_resource_group.rg > /dev/null;"