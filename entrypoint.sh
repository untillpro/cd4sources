#!/bin/bash

REPO=${REPO:-"https://github.com/untillpro/directcd-test"}
OUT=${OUT:-"out.exe"}
TIMEOUT=${TIMEOUT:-"10"}
WORKING_FOLDER=${WORKING_FOLDER:-"/directcd/tmp"}
VERBOSE=${VERBOSE:-"false"}
REPLACE=${REPLACE:-""}
ARGS=${ARGS:-""}

exec /directcd/directcd pull --repo=${REPO} -o=${OUT} -t=${TIMEOUT} -w=${WORKING_FOLDER} -v=${VERBOSE} --replace=${REPLACE} -- ${ARGS}