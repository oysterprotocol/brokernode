#!/bin/bash

chmod 600 ./travis/id_rsa

read -r -d '' DEPLOY_SCRIPT << EOM
  sudo su;
  cd /home/ubuntu/brokernode;
  git stash;
  git clean -f -d;
  git checkout master;
  git pull;
  cp /home/ubuntu/database.dev.yml ./database.yml;
  cp /home/ubuntu/docker-compose.dev.yml ./docker-compose.yml;
  docker stop $(docker ps -aq);
  docker rm $(docker ps -aq);
  docker rmi $(docker images -q);
  DEBUG=1 docker-compose up --build -d;
EOM

TEST_BROKERS=("52.14.218.135" "18.217.133.146")

for ip_address in "${TEST_BROKERS[@]}"
do
  ssh -o StrictHostKeyChecking=no ubuntu@$ip_address -i ./travis/id_rsa << END
    $DEPLOY_SCRIPT
END
done
