#!/bin/bash

echo "HEREEEEEEEEEEEEEEEE"
pwd

chmod 600 ./travis/id_rsa

script=$(cat <<-END
  echo "PRINT SOMETHINGGGGGGGGGGGGGGGGG"
  cd brokernode
END
)

ssh -o StrictHostKeyChecking=no ubuntu@52.14.218.135 -i ./travis/id_rsa $script
