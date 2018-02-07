#!/bin/bash
echo "Starting"
git add $(git rev-parse --show-toplevel)
git commit -a 
git checkout -b host
git push -u origin host
ssh oyster@192.168.56.20 'cd brokernode && git merge master host && git checkout master && git branch -D host'
#change host if you want another branch name
git checkout master
git pull origin master
git branch -D host
echo "Done! This window will close in 10 seconds, just in case you need to look at any errors"
sleep 30