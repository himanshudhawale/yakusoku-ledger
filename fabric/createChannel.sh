echo "........Creating channel for Clemson University........"
export CORE_PEER_ADDRESS=peer0.university.clemson.com:7051
export CORE_PEER_LOCALMSPID=UniversityMSP
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/channel-artifacts/certs/peerOrganizations/university.clemson.com/users/Admin@university.clemson.com/msp

peer channel create -o orderer.clemson.com:7050 -c channel1 -f ./channel-artifacts/channel/channel.tx --tls --cafile /opt/gopath/channel-artifacts/certs/ordererOrganizations/clemson.com/orderers/orderer.clemson.com/tls/ca.crt

sleep 20

echo ".......Joining peers of Clemson........."
peer channel join -b channel1.block

export CORE_PEER_ADDRESS=peer1.university.clemson.com:7051
peer channel join -b channel1.block

export CORE_PEER_LOCALMSPID=StudentMSP
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/channel-artifacts/certs/peerOrganizations/student.clemson.com/users/Admin@student.clemson.com/msp
export CORE_PEER_ADDRESS=peer0.student.clemson.com:7051
peer channel join -b channel1.block

export CORE_PEER_ADDRESS=peer1.student.clemson.com:7051
peer channel join -b channel1.block



echo "........Updating anchor peers........."
export CORE_PEER_LOCALMSPID=UniversityMSP
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/channel-artifacts/certs/peerOrganizations/university.clemson.com/users/Admin@university.clemson.com/msp
export CORE_PEER_ADDRESS=peer0.university.clemson.com:7051
peer channel update -o orderer.clemson.com:7050 -c channel1 -f ./channel-artifacts/channel/UniversityMSPanchors.tx --tls --cafile /opt/gopath/channel-artifacts/certs/ordererOrganizations/clemson.com/orderers/orderer.clemson.com/tls/ca.crt

export CORE_PEER_LOCALMSPID=StudentMSP
export CORE_PEER_MSPCONFIGPATH=/opt/gopath/channel-artifacts/certs/peerOrganizations/student.clemson.com/users/Admin@student.clemson.com/msp
export CORE_PEER_ADDRESS=peer0.student.clemson.com:7051
peer channel update -o orderer.clemson.com:7050 -c channel1 -f ./channel-artifacts/channel/StudentMSPanchors.tx --tls --cafile /opt/gopath/channel-artifacts/certs/ordererOrganizations/clemson.com/orderers/orderer.clemson.com/tls/ca.crt
