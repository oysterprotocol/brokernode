#!/bin/bash

echo "HEREEEEEEEEEEEEEEEE"
pwd

chmod 600 ./travis/id_rsa

ssh -o StrictHostKeyChecking=no ubuntu@52.14.218.135 -i ./travis/id_rsa <<EOF
  echo "PRINT SOMETHINGGGGGGGGGGGGGGGGG"
EOF
