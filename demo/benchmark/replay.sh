#!/bin/bash

# This script will launch a cluster of Lachesis nodes
# The parameter N = number of nodes to run

set -e

# number of nodes N
N=5
LIMIT_CPU=$(echo "scale=2; 1/$N" | bc)
LIMIT_IO=$(echo "500/$N" | bc)

#
PROG=lachesis
EXEC=../../build/lachesis

# default ip using localhost
IP=127.0.0.1
# default port PORT
# the actual ports are PORT+1, PORT+2, etc (18541, 18542, 18543, ... )
PORT=18540

LACHESIS_BASE_DIR=/tmp/lachesis-demo-replay/datadir
rm -rf ${LACHESIS_BASE_DIR} > /dev/null
mkdir ${LACHESIS_BASE_DIR}

declare -a pids

# Set up plg tool
cp ${GOPATH}/src/github.com/Fantom-foundation/tun-replay/build/plg .
PLG=./plg
#echo "PLG=${PLG}"


# Function to be called when user press ctrl-c.
ctrl_c() {
    echo "Shell terminating ..."
    for pid in "${pids[@]}"
    do
	echo "Killing ${pid}..."
	kill -9 ${pid}
	# Suppresses "[pid] Terminated ..." message.
	wait ${pid} &>/dev/null
    done
    exit;
}

trap ctrl_c SIGHUP SIGINT
echo "You may press ctrl-c to kill all started processes..."

declare -a progs

echo -e "\nStart $N nodes:\n"
for i in $(seq $N)
do
    port=$((PORT + i))
    localport=$((5050 + i))

    prog="exec${i}.sh"

    echo "#!/bin/bash" > ${prog}
    echo "${EXEC} --fakenet $i/$N --port ${localport} \
      --rpc --rpcapi \"eth,ftm,debug,admin,web3\" --rpcport ${port} --nousb --verbosity 3 \
      --datadir ${LACHESIS_BASE_DIR}/lach$i" >> ${prog}
    chmod +x ${prog}
    progs+="\"./${prog}\" "
done

finalcmd="sudo ${PLG} --record --subnet 127.0.0.1/24 -f dump.traffic  ${progs}"
echo -e "Started replaying"
echo -e "${finalcmd}"

# execute the plg command
# sudo ${PLG} --record --subnet 10.0.0.0/24 -f dump.traffic  ${progs}
sudo ${finalcmd}


# now connecting the nodes

attach_and_exec() {
     local URL=$1
     local CMD=$2

     for attempt in $(seq 20)
     do
         if (( attempt > 5 ));
         then
             echo "  - attempt ${attempt}: " >&2
         fi;

         res=$("${EXEC}" --exec "${CMD}" attach http://${URL} 2> /dev/null)
         if [ $? -eq 0 ]
         then
             #echo "success" >&2
             echo $res
             return 0
         else
             #echo "wait" >&2
             sleep 1
         fi
     done
     echo "failed RPC connection to ${NAME}" >&2
     return 1
 }


 echo -e "\nConnect nodes to ring:\n"
 for i in $(seq $N)
 do
     j=$((i % N + 1))

     echo " getting node-$j address:"
 	 url=${IP}:$((PORT + j))
 	 echo "    at url: ${url}"

     enode=$(attach_and_exec ${url} 'admin.nodeInfo.enode')
     echo "    p2p address = ${enode}"

     echo " connecting node-$i to node-$j:"
     url=${IP}:$((PORT + i))
     echo "    at url: ${url}"

     res=$(attach_and_exec ${url} "admin.addPeer(${enode})")
     echo "    result = ${res}"
 done



echo
echo "Sleep for 10000 seconds..."
sleep 10000
