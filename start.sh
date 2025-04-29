#!/bin/sh

DEFAULT_USER_PASS=$(./encrypt $DEFAULT_USER_PASS)

exec "$@"