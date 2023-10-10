#!/usr/bin/env bash

export VAULT_ADDR=http://localhost:8200

checkResult() {
    result=$(echo $?)
    if [[ $result -eq 0 ]]; then
        echo -e "\n$1\n"
    else
        echo "$2"
        exit 1
    fi
}

echo -e "\n################### - Login to vault - #####################\n"

vault login -method=userpass username=local-user password=local-pwd
checkResult "Login successful" "Login failed"

echo -e "############################################################\n"

echo -e "################# - create a new service - #################\n"

vault write ethereum/key-managers serviceName="my-service"
checkResult "Service created" "Service creation failed"

echo -e "############################################################\n"

echo -e "################## - list the services - ###################\n"

vault list ethereum/key-managers 
checkResult "Service listed" "Service listing failed"

echo -e "############################################################\n"

echo -e "################## - read the service - ####################\n"

vault read ethereum/key-managers/my-service
checkResult "Service read" "Service read failed"

echo -e "############################################################\n"

echo -e "################ - delete the service - ####################\n"

vault delete ethereum/key-managers/my-service
checkResult "Service deleted" "Service deletion failed"

echo -e "############################################################\n"

echo "test completed successfully"