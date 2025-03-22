#!/bin/bash

# The use is 'bash deploy.sh maze-wars' if you want to deploy 'maze-wars'
# service or any other of the services on the production.yml, if you
# just want to update the 'data/' then just run
# the deploy empty and only the file will be uploaded

set -e

# Uses rancher-machine to deploy and the machine is named 'maze-wars'
eval $(rancher-machine env maze-wars)

# On the server this is the default folder so this is where it'll land
# and this is the default one that the docker will try to read the data from.
rancher-machine scp -r docker/data/ maze-wars:/home/maze-wars

# It'll deploy all the services passed as parameter, if none
# is passed no deploy of services will be done
if (( $# != 0 )); then
    docker-compose -f docker/docker-compose.yml -f docker/production.yml up --build --force-recreate --remove-orphans --no-deps -d $@
fi

# Remove old images so we do not keep them if they are not needed
docker image prune -a -f --filter "until=24h"

# Not needed as the script ends and it's exited, but I leave it here
# just for documentation sake
# rancher-machine env --unset
