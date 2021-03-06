package main

import (
	"bufio"
	"fmt"
	"log"

	pbmsg "github.com/ethresearch/sharding-p2p-poc/pb/message"
	inet "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	ma "github.com/multiformats/go-multiaddr"
	protobufCodec "github.com/multiformats/go-multicodec/protobuf"
	"golang.org/x/crypto/sha3"

	"github.com/golang/protobuf/proto"
	b58 "github.com/mr-tron/base58/base58"
)

func keccak(msg proto.Message) []byte {
	dataInBytes, err := proto.Marshal(msg)
	if err != nil {
		log.Printf("Error occurs when hashing %v", msg)
	}
	h := sha3.NewLegacyKeccak256()
	h.Write(dataInBytes)
	return h.Sum(nil)
}

func GetCollationHash(msg *pbmsg.Collation) string {
	hashBytes := keccak(msg)
	// FIXME: for convenience
	return b58.Encode(hashBytes)
}

// helper method - writes a protobuf go data object to a network stream
// data: reference of protobuf go data object to send (not the object itself)
// s: network stream to write the data to
func sendProtoMessage(data proto.Message, s inet.Stream) bool {
	writer := bufio.NewWriter(s)
	enc := protobufCodec.Multicodec(nil).Encoder(writer)
	err := enc.Encode(data)
	if err != nil {
		log.Println(err)
		return false
	}
	writer.Flush()
	return true
}

func parseAddr(addrString string) (peerID peer.ID, protocolAddr ma.Multiaddr) {
	// The following code extracts target's the peer ID from the
	// given multiaddress
	ipfsaddr, err := ma.NewMultiaddr(addrString) // ipfsaddr=/ip4/127.0.0.1/tcp/10000/ipfs/QmVmDaabYcS3pn23KaFjkdw6hkReUUma8sBKqSDHrPYPd2
	if err != nil {
		log.Fatalln(err)
	}

	pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS) // pid=QmVmDaabYcS3pn23KaFjkdw6hkReUUma8sBKqSDHrPYPd2
	if err != nil {
		log.Fatalln(err)
	}

	peerid, err := peer.IDB58Decode(pid) // peerid=<peer.ID VmDaab>
	if err != nil {
		log.Fatalln(err)
	}

	// Decapsulate the /ipfs/<peerID> part from the target
	// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
	targetPeerAddr, _ := ma.NewMultiaddr(
		fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)),
	)
	targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)
	return peerid, targetAddr
}
