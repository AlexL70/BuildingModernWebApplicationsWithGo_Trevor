#!/bin/bash
# take the latest version from the repo
git pull

# move to the app directory
cd ./bookings
# run migration on postgresql db
soda migrate
# rebuild app
go build -o bookings ./cmd/web
# move back to git root directory
cd ./..
# restore schema dump after migration (differs because of different OS)
git checkout bookings/migrations/schema.sql
# stop bookings web app
sudo supervisorctl stop bookings
# restart bookings web app
sudo supervisorctl start bookings
