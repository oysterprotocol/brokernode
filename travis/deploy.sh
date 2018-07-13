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
  sudo curl https://gist.githubusercontent.com/mlebkowski/471d2731176fb11e81aa/raw/4a6a18a19ac5a9f1b8c6533a07f93e01da8ddba0/cleanup-docker.sh | bash;
  real_ip=$(curl icanhazip.com )
  sudo sed -i -e 's/127.0.0.1/$real_ip/' ./.env
  DEBUG=1 docker-compose up --build -d;
EOM

TRAVIS_BROKERS=("18.222.56.121" "18.191.77.193")

for ip_address in "${TRAVIS_BROKERS[@]}"
do
  ssh -o StrictHostKeyChecking=no ubuntu@$ip_address -i ./travis/id_rsa << END
    $DEPLOY_SCRIPT
END
done
