#!/usr/bin/env bash

declare -ri n="${n:-3}"

file=blockade.yaml

echo "containers:" > $file

for i in $(seq $n)
do
    j=$((i % n + 1)) # ring


echo "  node-$i:
    image: pos-lachesis-blockade:latest
    name: node-$i
    hostname: node-$i
    command: start --fakegen=$i/$n --db=/tmp --peer=blockade_node-$j
    expose:
      - \"55555\"
      - \"55556\"
      - \"55557\"" >> $file

if [ "$i" -ne "1" ]; then 
echo "    links: [\"node-$j\"]
" >> $file
fi

done