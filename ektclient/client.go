package ektclient

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"strconv"

	"github.com/EducationEKT/EKT/blockchain"
	"github.com/EducationEKT/EKT/core/types"
	"github.com/EducationEKT/EKT/crypto"
	"github.com/EducationEKT/EKT/util"

	"github.com/EducationEKT/xserver/x_http/x_resp"
)

type IClient interface {
	// block
	GetHeaderByHeight(height int64) *blockchain.Header
	GetBlockByHeight(height int64) *blockchain.Block
	GetLastBlock(peer types.Peer) *blockchain.Header
	GetHeaderByHash(hash []byte) *blockchain.Header

	// vote
	GetVotesByBlockHash(hash string) blockchain.Votes

	// delegate
	BroadcastBlock(block blockchain.Block)
	SendVote(vote blockchain.PeerBlockVote)
	SendVoteResult(votes blockchain.Votes)
}

type Client struct {
	peers []types.Peer
}

func NewClient(peers []types.Peer) IClient {
	return Client{peers: peers}
}

func (client Client) GetHeaderByHeight(height int64) *blockchain.Header {
	for _, peer := range client.peers {
		url := util.StringJoint("http://", peer.Address, ":", strconv.Itoa(int(peer.Port)), "/block/api/getHeaderByHeight?height=", strconv.Itoa(int(height)))
		body, err := util.HttpGet(url)
		if err != nil {
			continue
		}
		if header := blockchain.FromBytes2Header(body); header != nil {
			return header
		}
	}
	return nil
}

func (client Client) GetBlockByHeight(height int64) *blockchain.Block {
	for _, peer := range client.peers {
		url := util.StringJoint("http://", peer.Address, ":", strconv.Itoa(int(peer.Port)), "/block/api/getBlockByHeight?height=", strconv.Itoa(int(height)))
		body, err := util.HttpGet(url)
		if err != nil {
			continue
		}
		if block := blockchain.GetBlockFromBytes(body); block == nil || len(block.Hash) == 0 {
			continue
		} else {
			return block
		}
	}
	return nil
}

func (client Client) GetHeaderByHash(hash []byte) *blockchain.Header {
	for _, peer := range client.peers {
		data, err := peer.GetDBValue(hex.EncodeToString(hash))
		if err == nil && bytes.EqualFold(crypto.Sha3_256(data), hash) {
			var header blockchain.Header
			err := json.Unmarshal(data, &header)
			if err == nil {
				return &header
			}
		}
	}
	return nil
}

func (client Client) GetLastBlock(peer types.Peer) *blockchain.Header {
	for _, peer := range client.peers {
		url := util.StringJoint("http://", peer.Address, ":", strconv.Itoa(int(peer.Port)), "/block/api/last")
		body, err := util.HttpGet(url)
		if err != nil {
			continue
		}
		if block := blockchain.FromBytes2Header(body); block != nil {
			return block
		}
	}
	return nil
}

func (client Client) GetVotesByBlockHash(hash string) blockchain.Votes {
	for _, peer := range client.peers {
		url := util.StringJoint("http://", peer.Address, ":", strconv.Itoa(int(peer.Port)), "/vote/api/getVotes?hash=", hash)
		body, err := util.HttpGet(url)
		if err != nil {
			continue
		}
		var result x_resp.XRespBody
		err = json.Unmarshal(body, &result)
		if err != nil || result.Status < 0 || result.Result == nil {
			continue
		}
		data, err := json.Marshal(result.Result)
		if err != nil {
			continue
		}
		if votes := GetVotesFromResp(data); len(votes) != 0 {
			return votes
		}
	}
	return nil
}

func (client Client) BroadcastBlock(block blockchain.Block) {
	data := block.Bytes()
	for _, peer := range client.peers {
		url := util.StringJoint("http://", peer.Address, ":", strconv.Itoa(int(peer.Port)), "/block/api/blockFromPeer")
		go util.HttpPost(url, data)
	}
}

func (client Client) SendVote(vote blockchain.PeerBlockVote) {
	data := vote.Bytes()
	for _, peer := range client.peers {
		url := util.StringJoint("http://", peer.Address, ":", strconv.Itoa(int(peer.Port)), "/vote/api/vote")
		go util.HttpPost(url, data)
	}
}

func (client Client) SendVoteResult(votes blockchain.Votes) {
	data := votes.Bytes()
	for _, peer := range client.peers {
		url := util.StringJoint("http://", peer.Address, ":", strconv.Itoa(int(peer.Port)), "/vote/api/voteResult")
		go util.HttpPost(url, data)
	}
}

func GetVotesFromResp(body []byte) blockchain.Votes {
	var votes blockchain.Votes
	json.Unmarshal(body, &votes)
	return votes
}
