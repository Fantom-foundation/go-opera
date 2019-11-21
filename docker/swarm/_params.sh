#!/usr/bin/env bash
. $(dirname $0)/../_params.sh

REGISTRY_HOST=192.168.88.10
SWARM_HOST=192.168.88.10

SSLDIR=$(dirname $0)/ssl
CA_FILE=`ls -1 ${SSLDIR}/*CA*.crt`
CERT_FILE=`ls -1 ${SSLDIR}/*.crt | grep -v CA`
KEY_FILE=`ls -1 ${SSLDIR}/*.key`

SWARM="--tls --tlscacert=${CA_FILE} --tlscert=${CERT_FILE} --tlskey=${KEY_FILE} -H tcp://${SWARM_HOST}:2376"
