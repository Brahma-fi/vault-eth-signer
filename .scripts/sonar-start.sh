#!/usr/bin/env bash

########################
# requirements:        #
# 1. docker            #
# 2. curl              #
# 3. jq                #
# 5. sonar-scanner     #
########################

set -e

user_tokenName="local_token"
username="admin"
user_password="admin"
new_password="1234"
url="http://localhost"
port="9000"

if [[ "$( docker container inspect -f '{{.State.Running}}' sonarqube )" == "true" ]];
then
  docker ps
else
  docker run --rm -d --name sonarqube -p 9000:9000 sonarqube
fi

echo "waiting for sonarqube starts..."
curl -s "$@" http://localhost:9000 | awk '/STARTING/{ print $0 }' | xargs

STATUS="$(curl -s "$@" http://localhost:9000 | awk '/UP/{ print $0 }')"
while [ -z "$STATUS" ]
do
	sleep 2
	STATUS="$(curl -s "$@" http://localhost:9000 | awk '/UP/{ print $0 }')"
	printf "."
done

printf '\n %s' "${STATUS}" | xargs
echo ""

# change the default password to avoid create a new one when login for the very first time
curl -u ${username}:${user_password} -X POST "${url}:${port}/api/users/change_password?login=${username}&previousPassword=${user_password}&password=${new_password}"

# search the specific user tokens for SonarQube
hasToken=$(curl --silent -u ${username}:${new_password} -X GET "${url}:${port}/api/user_tokens/search")
if [[ -n "${hasToken}"  ]]; then
  # Revoke the user token for SonarQube
  curl -X POST -H "Content-Type: application/x-www-form-urlencoded" -d "name=${user_tokenName}" -u ${username}:${new_password} "${url}:${port}"/api/user_tokens/revoke
fi

# generate new token
token=$(curl --silent -X POST -H "Content-Type: application/x-www-form-urlencoded" -d "name=${user_tokenName}" -u ${username}:${new_password} "${url}:${port}"/api/user_tokens/generate | jq '.token' | xargs)

# scan and push the results to localhost docker container
sonar-scanner -D sonar.projectKey="vault-eth-signer" \
              -D sonar.projectName="vault-eth-signer" \
              -D sonar.scm.provider=git \
              -D sonar.sources=. \
              -D sonar.exclusions=".bin/**,**/*_test.go" \
              -D sonar.tests=./ \
              -D sonar.test.inclusions=./**/*_test.go \
              -D sonar.go.tests.reportPaths=.report/report.out \
              -D sonar.go.coverage.reportPaths=.coverage/*.out \
              -D sonar.host.url="${url}:${port}" \
              -D sonar.github.repository='https://github.com/Brahma-fi/vault-eth-signer' \
              -D sonar.token="${token}"
