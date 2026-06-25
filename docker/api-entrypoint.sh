#!/bin/sh
set -e

./migrator -up
exec ./server
