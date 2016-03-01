#!/bin/sh

/mumble-authenticator/env.sh
python3 /mumble-authenticator/auth.py &
/go/bin/fumble
