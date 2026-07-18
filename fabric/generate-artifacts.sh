#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")"

rm -rf channel-artifacts/certs channel-artifacts/channel
mkdir -p channel-artifacts/channel

./cryptogen generate \
	--config=config/crypto-config.yaml \
	--output=channel-artifacts/certs

export FABRIC_CFG_PATH="$PWD/config"
./configtxgen \
	-profile GenesisBlockGeneration \
	-channelID system-channel \
	-outputBlock channel-artifacts/channel/genesis.block
./configtxgen \
	-profile ChannelCreation \
	-channelID channel1 \
	-outputCreateChannelTx channel-artifacts/channel/channel.tx
./configtxgen \
	-profile ChannelCreation \
	-channelID channel1 \
	-asOrg UniversityMSP \
	-outputAnchorPeersUpdate channel-artifacts/channel/UniversityMSPanchors.tx
./configtxgen \
	-profile ChannelCreation \
	-channelID channel1 \
	-asOrg StudentMSP \
	-outputAnchorPeersUpdate channel-artifacts/channel/StudentMSPanchors.tx

echo "Generated Fabric crypto material and channel artifacts."
