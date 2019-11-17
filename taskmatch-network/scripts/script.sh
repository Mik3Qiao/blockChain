#!/bin/bash

echo
echo " ____    _____      _      ____    _____ "
echo "/ ___|  |_   _|    / \    |  _ \  |_   _|"
echo "\___ \    | |     / _ \   | |_) |   | |  "
echo " ___) |   | |    / ___ \  |  _ <    | |  "
echo "|____/    |_|   /_/   \_\ |_| \_\   |_|  "
echo
echo "Taskmatching network being started!"
echo
CHANNEL_NAME="$1"
DELAY="$2"
LANGUAGE="$3"
TIMEOUT="$4"
VERBOSE="$5"
NO_CHAINCODE="$6"
: ${CHANNEL_NAME:="taskmatch-channel"}
: ${DELAY:="3"}
: ${LANGUAGE:="golang"}
: ${TIMEOUT:="10"}
: ${VERBOSE:="false"}
: ${NO_CHAINCODE:="false"}
LANGUAGE=`echo "$LANGUAGE" | tr [:upper:] [:lower:]`
COUNTER=1
MAX_RETRY=10

CC_SRC_PATH="github.com/chaincode/taskmatching/"

echo "Channel name : "$CHANNEL_NAME

# import utils, some variables are also defined in here.
. scripts/utils.sh

## Creates the channel
createChannel() {
	setGlobals 0 1

	set -x
	peer channel create -o orderer.example.com:7050 -c $CHANNEL_NAME -f ./channel-artifacts/channel.tx --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA >&log.txt
	res=$?
	set +x

	cat log.txt
	verifyResult $res "Channel creation failed"
	echo "===================== Channel '$CHANNEL_NAME' created ===================== "
	echo
}

## Joins peers to the channel
joinChannel () {
	for org in 1 2 3; do
		joinChannelWithRetry 0 $org
		echo "===================== peer0.org${org} joined channel '$CHANNEL_NAME' ===================== "
		sleep $DELAY
		echo
	done
}

## Create channel
echo "Creating channel..."
createChannel

## Join all the peers to the channel
echo "Having all peers join the channel..."
joinChannel



## Update the anchor peers for each org in the channel
echo "Updating anchor peers for org1..."
updateAnchorPeers 0 1

echo "Updating anchor peers for org2..."
updateAnchorPeers 0 2

echo "Updating anchor peers for org3..."
updateAnchorPeers 0 3

## Install the chaincode for each peer
echo "Installing chaincode on peer0.org1..."
installChaincode 0 1
echo "Install chaincode on peer0.org2..."
installChaincode 0 2
echo "Install chaincode on peer0.org3..."
installChaincode 0 3

## Instantiate chaincode on peer0.org2, it only needs to be instantiated on one peer.
echo "Instantiating chaincode on peer0.org2..."
instantiateChaincode 0 2

echo "Sleeping for 3 seconds to update ledger..."
sleep 3

 
echo "Invoking chaincode:"
chaincodeInvoke ## Comment out this line if you don't want the example to be run!

echo
echo "========= All Tests Completed Successfully =========== "
echo

echo
echo " _____   _   _   ____   "
echo "| ____| | \ | | |  _ \  "
echo "|  _|   |  \| | | | | | "
echo "| |___  | |\  | | |_| | "
echo "|_____| |_| \_| |____/  "
echo

exit 0
