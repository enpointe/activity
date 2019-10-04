#!/bin/bash -x

# Determine the absolute path for the top of the development tree
# (workspace) to use for mounting into docker
WORKSPACE="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# The default name to assign to the docker container
# This can be changed via the -c command line option
CONTAINER_NAME=activity


start() {

  # Working directory inside the container.
  local container_workdir=/go/src/github.com/enpointe/activity

  docker run --rm -it \
    --name $CONTAINER_NAME \
    --volume $WORKSPACE:$container_workdir \
    --workdir $container_workdir \
    golang
}

help() {
    local progname=$(basename $0)
    cat << EOF
Usage: $progname [OPTIONS]

A startup script used to setup a docker container that can be used
for developing, maintaining, and testing this project

Options:
    -c  Name to give to docker container, defaults to "$CONTAINER_NAME"
    -h  Usage help

EOF
}

while getopts c:h option
do
    case "${option}"
    in
    c) CONTAINER_NAME=${OPTARG}
       echo "Using specified container name of $CONTAINER_NAME";;
    h) help;;
esac
done

start
