#!/bin/sh

cd /mumble-authenticator/
./env.sh
python3 auth.py &
/go/bin/fumble
