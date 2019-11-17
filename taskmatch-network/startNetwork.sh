export PATH=${PWD}/../bin:${PWD}:$PATH
export FABRIC_CFG_PATH=${PWD}
export VERBOSE=false

# Obtain the OS and Architecture string that will be used to select the correct
# native binaries for your platform, e.g., darwin-amd64 or linux-amd64
OS_ARCH=$(echo "$(uname -s | tr '[:upper:]' '[:lower:]' | sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')" | awk '{print tolower($0)}')
# timeout duration - the duration the CLI should wait for a response from
# another container before giving up
CLI_TIMEOUT=10
# default for delay between commands
CLI_DELAY=3
# system channel name defaults to "byfn-sys-channel"
SYS_CHANNEL="taskmatch-sys-channel"
# channel name defaults to "mychannel"
CHANNEL_NAME="taskmatch-channel"
# use this as the default docker-compose yaml definition
COMPOSE_FILE=docker-compose-cli.yaml

# use golang as the default language for chaincode
LANGUAGE=golang
# default image tag
IMAGETAG="latest"
# default consensus type
CONSENSUS_TYPE="solo"

echo
echo "##########################################################"
echo "##### Generate certificates using cryptogen tool #########"
echo "##########################################################"
echo 


# remove orderer block and other channel configuration transactions and certs
rm -rf channel-artifacts/*.block channel-artifacts/*.tx crypto-config

set -x
cryptogen generate --config=./crypto-config.yaml --output="crypto-config"
res=$?
set +x

if [ $res -ne 0 ]; then
    echo "Failed to generate certificates..."
    exit 1
fi

echo 
echo "##########################################################"
echo "#########  Generating Orderer Genesis block ##############"
echo "##########################################################"
echo 

mkdir channel-artifacts

set -x
configtxgen -profile ThreeOrgsOrdererGenesis -channelID $SYS_CHANNEL -outputBlock ./channel-artifacts/genesis.block
res=$?
set +x

set +x
if [ $res -ne 0 ]; then
    echo "Failed to generate orderer genesis block..."
    exit 1
fi

echo
echo "#################################################################"
echo "### Generating channel configuration transaction 'channel.tx' ###"
echo "#################################################################"
set -x
configtxgen -profile ThreeOrgsChannel -outputCreateChannelTx ./channel-artifacts/channel.tx -channelID $CHANNEL_NAME
res=$?
set +x

if [ $res -ne 0 ]; then
    echo "Failed to generate channel configuration transaction..."
    exit 1
fi

echo
echo "#################################################################"
echo "#######    Generating anchor peer update for Org1MSP   ##########"
echo "#################################################################"
set -x
configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate ./channel-artifacts/Org1MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org1MSP
res=$?
set +x

if [ $res -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org1MSP..."
    exit 1
fi

echo
echo "#################################################################"
echo "#######    Generating anchor peer update for Org2MSP   ##########"
echo "#################################################################"
set -x
configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate \
./channel-artifacts/Org2MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org2MSP
res=$?
set +x

if [ $res -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org2MSP..."
    exit 1
fi

echo
echo "#################################################################"
echo "#######    Generating anchor peer update for Org3MSP   ##########"
echo "#################################################################"
set -x
configtxgen -profile ThreeOrgsChannel -outputAnchorPeersUpdate \
./channel-artifacts/Org3MSPanchors.tx -channelID $CHANNEL_NAME -asOrg Org3MSP
res=$?
set +x

if [ $res -ne 0 ]; then
    echo "Failed to generate anchor peer update for Org2MSP..."
    exit 1
fi


echo 
echo "#################################################################"
echo "##  Shutting down any docker containers used by this network  ###"
echo "#################################################################"
echo 
docker-compose -f docker-compose-cli.yaml down

echo 
echo "#################################################################"
echo "#######           Starting up docker containers        ##########"
echo "#################################################################"
echo 
docker-compose -f docker-compose-cli.yaml up -d 2>&1 #2>&1 suppresses output from containers, comment it out if you're curious to see what happens.
echo "passed here."

## Begin installing and instantiating chaincode onto the ledger
docker exec cli scripts/script.sh $CHANNEL_NAME $CLI_DELAY $LANGUAGE $CLI_TIMEOUT $VERBOSE $NO_CHAINCODE