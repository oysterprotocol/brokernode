#!/bin/bash

echo "HEREEEEEEEEEEEEEEEE"
pwd

chmod 600 ./travis/id_rsa

script=$(cat <<-END
  sudo su;
  cd /home/ubuntu/brokernode;
  git stash;
  git pull;
  cp /home/ubuntu/database.dev.yml ./database.yml;
  cp /home/ubuntu/docker-compose.dev.yml ./docker-compose.yml;
  docker stop $(docker ps -aq);
  docker rm $(docker ps -aq);
  docker rmi $(docker images -q);
  DEBUG=1 docker-compose up --build -d;
END
)

ssh -o StrictHostKeyChecking=no ubuntu@52.14.218.135 -i ./travis/id_rsa $script
