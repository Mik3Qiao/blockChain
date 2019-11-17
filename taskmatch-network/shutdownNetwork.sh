echo 
echo "#################################################################"
echo "##################  Shutting Down The Network  ##################"
echo "#################################################################"
echo 
docker-compose -f docker-compose-cli.yaml down --volumes --remove-orphans
rm -rf channel-artifacts crypto-config
