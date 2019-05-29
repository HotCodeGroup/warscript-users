#!/bin/bash

set -e

echo -e "# Building docker.\n"
docker build -t warscript-users .
docker tag warscript-users $DOCKER_USER/warscript-users
docker push $DOCKER_USER/warscript-users

echo -e "# Starting docker.\n"
echo $CONSUL_ADDR

chmod 600 ./2019_1_HotCode_id_rsa.pem
ssh-keyscan -H 89.208.198.192 >> ~/.ssh/known_hosts
for (( c=1; c<=$CONTAINERS_COUNT; c++ ))
do
    ssh -i ./2019_1_HotCode_id_rsa.pem ubuntu@89.208.198.192 docker stop warscript-users.$c
    if ssh -i ./2019_1_HotCode_id_rsa.pem ubuntu@89.208.198.192 test $? -eq 0
    then
        ssh -i ./2019_1_HotCode_id_rsa.pem ubuntu@89.208.198.192 docker rm warscript-users.$c || true
    fi
    ssh -i ./2019_1_HotCode_id_rsa.pem ubuntu@89.208.198.192 docker run -e CONSUL_ADDR=$CONSUL_ADDR \
                                                                    -e VAULT_ADDR=$VAULT_ADDR \
                                                                    -e VAULT_TOKEN=$VAULT_TOKEN \
                                                                    --name=warscript-users.$c \
                                                                    -d --net=host $DOCKER_USER/warscript-users
done