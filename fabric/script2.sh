export FABRIC_CFG_PATH=./config && export NAME_OF_CHANNEL=channel1

./configtxgen -profile GenesisBlockGeneration -outputBlock ./channel-artifacts/channel/genesis.block


./configtxgen -profile ChannelCreation -outputCreateChannelTx ./channel-artifacts/channel/channel.tx -channelID $NAME_OF_CHANNEL
./configtxgen -profile ChannelCreation -outputAnchorPeersUpdate ./channel-artifacts/channel/UniversityMSPanchors.tx -channelID $NAME_OF_CHANNEL -asOrg UniversityMSP
./configtxgen -profile ChannelCreation -outputAnchorPeersUpdate ./channel-artifacts/channel/StudentMSPanchors.tx -channelID $NAME_OF_CHANNEL -asOrg StudentMSP
