#!/usr/bin/env bash
. ../_params.sh

REGISTRY_HOST=18.189.195.64
SWARM_HOST=18.189.195.64

SSLDIR=./ssl
CA_FILE=`ls -1 ${SSLDIR}/*CA*.crt`
CERT_FILE=`ls -1 ${SSLDIR}/*.crt | grep -v CA`
KEY_FILE=`ls -1 ${SSLDIR}/*.key`

SWARM="--tls --tlscacert=${CA_FILE} --tlscert=${CERT_FILE} --tlskey=${KEY_FILE} -H tcp://${SWARM_HOST}:2376"
