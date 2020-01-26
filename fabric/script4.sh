#!/bin/bash

CHANNEL_NAME="channel1"
LANGUAGE="golang"
TIMEOUT="10"
DELAY="3"
ORDERER_CA=/opt/gopath/channel-artifacts/certs/ordererOrganizations/clemson.com/orderers/orderer.clemson.com/msp/tlscacerts/tlsca.clemson.com-cert.pem


setOrdererGlobals() {
        CORE_PEER_LOCALMSPID="OrdererMSP"
        CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/channel-artifacts/certs/ordererOrganizations/clemson.com/orderers/orderer.clemson.com/msp/tlscacerts/tlsca.example.com-cert.pem
        CORE_PEER_MSPCONFIGPATH=/opt/gopath/channel-artifacts/certs/ordererOrganizations/clemson.com/users/Admin@example.com/msp
}

setGlobals () {
	PEER=$1
	ORG=$2
	if [ $ORG -eq 1 ] ; then
		 CORE_PEER_LOCALMSPID="UniversityMSP"
		 CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/channel-artifacts/certs/peerOrganizations/university.clemson.com/peers/peer0.university.clemson.com/tls/ca.crt
		 CORE_PEER_MSPCONFIGPATH=/opt/gopath/channel-artifacts/certs/peerOrganizations/university.clemson.com/users/Admin@university.clemson.com/msp
		if [ $PEER -eq 0 ]; then
			 CORE_PEER_ADDRESS=peer0.university.clemson.com:7051
		else
			 CORE_PEER_ADDRESS=peer1.university.clemson.com:7051
		fi
	elif [ $ORG -eq 2 ] ; then
		 CORE_PEER_LOCALMSPID="StudentMSP"
		 CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/channel-artifacts/certs/peerOrganizations/student.clemson.com/peers/peer0.student.clemson.com/tls/ca.crt
		 CORE_PEER_MSPCONFIGPATH=/opt/gopath/channel-artifacts/certs/peerOrganizations/student.clemson.com/users/Admin@student.clemson.com/msp
		if [ $PEER -eq 0 ]; then
			 CORE_PEER_ADDRESS=peer0.student.clemson.com:7051
		else
			 CORE_PEER_ADDRESS=peer1.student.clemson.com:7051
		fi

	else
		echo "================== ERROR !!! ORG Unknown =================="
	fi

	env |grep CORE
}


	verifyResult () {
		if [ $1 -ne 0 ] ; then
			echo "!!!!!!!!!!!!!!! "$2" !!!!!!!!!!!!!!!!"
	    echo "========= ERROR !!! FAILED to execute End-2-End Scenario ==========="
			echo
	   		exit 1
		fi
	}


installChaincode () {
	PEER=$1
	ORG=$2
  LANGUAGE="golang"
  CC_SRC_PATH="chaincode/"

	setGlobals $PEER $ORG
	VERSION=${3:-1.0}
        set -x
	peer chaincode install -n studentuniversity -v ${VERSION} -l ${LANGUAGE} -p ${CC_SRC_PATH} > log.txt  2>&1
	res=$?
        set +x
	cat log.txt
	verifyResult $res "Chaincode installation on peer${PEER}.${ORG} has Failed"
	echo "===================== Chaincode is installed on peer${PEER}.${ORG} ===================== "
	echo
}

instantiateChaincode () {
	PEER=$1
	ORG=$2
	setGlobals $PEER $ORG
	VERSION=${3:-1.0}


	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                set -x
		peer chaincode instantiate -o orderer.clemson.com:7050 -C $CHANNEL_NAME -n studentuniversity -l ${LANGUAGE} -v ${VERSION} -c '{"Args":["Init","studentA","universityA","12th July 2019","500", "110000012"]}' -P "OR	('studentMSP.peer','universityMSP.peer')" > log.txt  2>&1
		res=$?
                set +x
	else
                set -x
		peer chaincode instantiate -o orderer.clemson.com:7050 --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n studentuniversity -l ${LANGUAGE} -v 1.0 -c '{"Args":["Init","studentA","universityA","12th July 2019","500", "110000012"]}' -P "OR	('studentMSP.peer','universityMSP.peer')" > log.txt  2>&1
		res=$?
                set +x
	fi
	cat log.txt
	verifyResult $res "Chaincode instantiation on peer${PEER}.${ORG} on channel '$CHANNEL_NAME' failed"
	echo "===================== Chaincode Instantiation on peer${PEER}.${ORG} on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}



chaincodeQuery () {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG
  echo "===================== Querying on peer${PEER}.${ORG} on channel '$CHANNEL_NAME'... ===================== "
  local rc=1
  local starttime=$(date +%s)

  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep $DELAY
     echo "Attempting to Query peer${PEER}.${ORG} ...$(($(date +%s)-starttime)) secs"
     set -x
     peer chaincode query -C $CHANNEL_NAME -n studentuniversity -c '{"Args":["queryBuySellByCustomer","CustomerA"]}'  > log.txt  2>&1
	 res=$?
     set +x
     test $res -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
		 rc=0

  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on peer${PEER}.${ORG} on channel '$CHANNEL_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on peer${PEER}.${ORG} is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
	exit 1
  fi
}



chaincodeInvoke () {
	PEER=$1
	ORG=$2
	setGlobals $PEER $ORG

	echo "=============================================================="
	echo "===================== Invoke transaction ====================="
	echo "=============================================================="

	if [ -z "$CORE_PEER_TLS_ENABLED" -o "$CORE_PEER_TLS_ENABLED" = "false" ]; then
                set -x
		peer chaincode invoke -o orderer.clemson.com:7050 -C $CHANNEL_NAME -n studentuniversity -c '{"Args":["invokeFunctionBuySell","c953f3d49f60302bcdd1506fb10112ccae9564fa1fa985ad1372218aad32b48a"]}' > log.txt  2>&1
		res=$?
                set +x
	else
                set -x
		peer chaincode invoke -o orderer.clemson.com:7050  --tls $CORE_PEER_TLS_ENABLED --cafile $ORDERER_CA -C $CHANNEL_NAME -n studentuniversity -c '{"Args":["invokeFunctionBuySell","c953f3d49f60302bcdd1506fb10112ccae9564fa1fa985ad1372218aad32b48a"]}' > log.txt  2>&1
		res=$?
                set +x
	fi
	cat log.txt
	verifyResult $res "Invoke execution on peer${PEER}.org${ORG} failed "
	echo "===================== Invoke transaction on peer${PEER}.org${ORG} on channel '$CHANNEL_NAME' is successful ===================== "
	echo
}



chaincodeQueryHistory () {
  PEER=$1
  ORG=$2
  setGlobals $PEER $ORG
  echo "===================== Querying on peer${PEER}.${ORG} on channel '$CHANNEL_NAME'... ===================== "
  local rc=1
  local starttime=$(date +%s)
  while test "$(($(date +%s)-starttime))" -lt "$TIMEOUT" -a $rc -ne 0
  do
     sleep $DELAY
     echo "Attempting to Query peer${PEER}.${ORG} ...$(($(date +%s)-starttime)) secs"
     set -x
     peer chaincode query -C $CHANNEL_NAME -n studentuniversity -c '{"Args":["getHistoryForBuySell","studentA","universityA"]}'  > log.txt  2>&1
	 res=$?
     set +x
     test $res -eq 0 && VALUE=$(cat log.txt | awk '/Query Result/ {print $NF}')
		 rc=0

  done
  echo
  cat log.txt
  if test $rc -eq 0 ; then
	echo "===================== Query on peer${PEER}.${ORG} on channel '$CHANNEL_NAME' is successful ===================== "
  else
	echo "!!!!!!!!!!!!!!! Query result on peer${PEER}.${ORG} is INVALID !!!!!!!!!!!!!!!!"
        echo "================== ERROR !!! FAILED to execute End-2-End Scenario =================="
	echo
	exit 1
  fi
}




## Install chaincode on peer0.org1 and peer0.org2

echo "Installing chaincode on peer0.student..."
#installChaincode 0 2
echo "Install chaincode on peer0.university..."
#installChaincode 0 1


### Instantiate chaincode on peer0.org2
echo "Instantiating chaincode on peer0.university..."
#instantiateChaincode 0 1

# Query chaincode on peer0.org1
echo "Querying chaincode on peer0.university..."
chaincodeQuery 0 1

# Invoke chaincode on peer0.org1
echo "Sending invoke transaction on peer0.university..."
#chaincodeInvoke 0 1

# Query chaincode on peer0.org1
echo "Querying chaincode on peer0.university..."
#chaincodeQueryHistory 0 1
