#!/usr/bin/env bash

version=$1

# copy console-policy json config file represantion with dummy values to vault container
docker cp .scripts/default.hcl vault-server:/default.hcl

# if we have a file at ./.build/vault/plugins/vault-eth-signer, then we get the sha256 of the file
if [[ -f "./.build/vault/plugins/vault-eth-signer-$version" ]]; then
    sha256="$(shasum -a 256 ./.build/vault/plugins/vault-eth-signer-$version | cut -d' ' -f1)"
else
    echo "No plugin found, then no plugin will be registered."
fi

sleep 3
cat << EOF | docker exec --interactive vault-server sh

PLUGIN="/vault/plugins/vault-eth-signer-$version"
echo "Plugin file must exist at /vault/plugins/vault-eth-signer (and compiled to linux) in order to be registered"

# perform login
vault login root 

# enable kv secrets engine at path: apps
vault secrets enable -path=apps kv-v2

# create a policy called "default" with the permissions to read and write to the path: apps
vault policy write default default.hcl

# Enable userpass auth method and create a user called "local-user" with password "local-pwd"
vault auth enable userpass
vault write auth/userpass/users/local-user password=local-pwd policies="default"

# if the volume was mounted correctly and there is a plugin file, then register the plugin
if [[ -f $PLUGIN ]]; then
    cd /vault/plugins
    ls -lah
    echo "sha256: $sha256"
    vault plugin register -sha256="$sha256" -command="vault-eth-signer-$version" -version="$version" secret vault-eth-signer
    vault secrets enable -path=ethereum -description="Ethereum signer" -plugin-name=vault-eth-signer plugin
else
    echo "No plugin found, then no plugin was registered."
fi
EOF
