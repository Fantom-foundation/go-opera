#!/bin/bash

set -e

N=${1:-4}
DEST=${2:-"conf"}
IPBASE=${3:-10.0.1.}
PORT=${4:-1337}

for i in $(seq 1 $N)
do
	dest=$DEST/node$i
	mkdir -p $dest
	lachesis keygen | sed -n -e "2 w $dest/pub" -e "4,+4 w $dest/priv_key.pem"
	echo "$IPBASE$((9 +i)):$PORT" > $dest/addr
done

PFILE=$DEST/peers.json
echo "[" > $PFILE
for i in $(seq 1 $N)
do
	com=","
	if [[ $i == $N ]]; then
		com=""
	fi

	printf "\t{\n" >> $PFILE
	printf "\t\t\"NetAddr\":\"$(cat $DEST/node$i/addr)\",\n" >> $PFILE
	printf "\t\t\"PubKeyHex\":\"$(cat $DEST/node$i/pub)\"\n" >> $PFILE
	printf "\t}%s\n"  $com >> $PFILE

done
echo "]" >> $PFILE

for i in $(seq 1 $N)
do
	dest=$DEST/node$i
	cp $DEST/peers.json $dest/
	rm $dest/addr $dest/pub
done
