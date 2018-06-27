#!/bin/bash

echo "HEREEEEEEEEEEEEEEEE"
pwd

chmod 600 ./travis/id_rsa

script=$(cat <<-END
  cd brokernode;
  docker rmi $(docker images -q);
  DEBUG=1 docker-compose up --build -d;
END
)

ssh -o StrictHostKeyChecking=no ubuntu@52.14.218.135 -i ./travis/id_rsa $script
