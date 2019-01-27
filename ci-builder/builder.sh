#!/bin/sh

# $1 - argument number 1
GIT_URL=$1
BRANCH=$2
GCR_HOSTNAME=$3
TAG=$4
# TAG = [gcr hostname= [regional gcr hostname] + [project name]] + [repo name]

mkdir repo
git clone $GIT_URL repo
if [ $? -eq 0 ] # Successfull clone
then
cd repo 
git checkout $BRANCH
if [ $? -eq 0 ] # branch exists
then
docker build -t foobar .
if [ $? -eq 0 ] # build successfull
then
docker tag foobar $TAG
cat /opt/secrets/docker-login.json | docker login -u _json_key --password-stdin https://$GCR_HOSTNAME
#docker login -u _json_key -p "$(cat /opt/secrets/docker-login.json)" https://$GCR_HOSTNAME
docker push $TAG

else
exit $?
fi

else
exit $?
fi

else
exit $?
fi