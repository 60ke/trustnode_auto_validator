package main

import "time"

type TmResult struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		NodeInfo struct {
			ProtocolVersion struct {
				P2P   string `json:"p2p"`
				Block string `json:"block"`
				App   string `json:"app"`
			} `json:"protocol_version"`
			Id         string `json:"id"`
			ListenAddr string `json:"listen_addr"`
			Network    string `json:"network"`
			Version    string `json:"version"`
			Channels   string `json:"channels"`
			Moniker    string `json:"moniker"`
			Other      struct {
				TxIndex    string `json:"tx_index"`
				RpcAddress string `json:"rpc_address"`
			} `json:"other"`
		} `json:"node_info"`
		SyncInfo struct {
			LatestBlockHash     string    `json:"latest_block_hash"`
			LatestAppHash       string    `json:"latest_app_hash"`
			LatestBlockHeight   string    `json:"latest_block_height"`
			LatestBlockTime     time.Time `json:"latest_block_time"`
			EarliestBlockHash   string    `json:"earliest_block_hash"`
			EarliestAppHash     string    `json:"earliest_app_hash"`
			EarliestBlockHeight string    `json:"earliest_block_height"`
			EarliestBlockTime   time.Time `json:"earliest_block_time"`
			MaxPeerBlockHeight  string    `json:"max_peer_block_height"`
			CatchingUp          bool      `json:"catching_up"`
			TotalSyncedTime     string    `json:"total_synced_time"`
			RemainingTime       string    `json:"remaining_time"`
			TotalSnapshots      string    `json:"total_snapshots"`
			ChunkProcessAvgTime string    `json:"chunk_process_avg_time"`
			SnapshotHeight      string    `json:"snapshot_height"`
			SnapshotChunksCount string    `json:"snapshot_chunks_count"`
			SnapshotChunksTotal string    `json:"snapshot_chunks_total"`
			BackfilledBlocks    string    `json:"backfilled_blocks"`
			BackfillBlocksTotal string    `json:"backfill_blocks_total"`
		} `json:"sync_info"`
		ValidatorInfo struct {
			Address string `json:"address"`
			PubKey  struct {
				Type  string `json:"type"`
				Value string `json:"value"`
			} `json:"pub_key"`
			VotingPower string `json:"voting_power"`
		} `json:"validator_info"`
	} `json:"result"`
}

type TmPub struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  struct {
		Address string `json:"address"`
		PubKey  struct {
			Type  string `json:"type"`
			Value string `json:"value"`
		} `json:"pub_key"`
		VotingPower string `json:"voting_power"`
	} `json:"result"`
}

type BTM struct {
	BscNodePubkeyAddr string `json:"bsc_node_pubkey_addr"`
	TmNodePubkeyAddr  string `json:"tm_node_pubkey_addr"`
}

type BscTmMap struct {
	BscPubkeyAddrMaps []BTM `json:"bscTMPubkeyPairs"`
	WebPort           int   `json:"WebPort"`
}

type BscResult struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Result  string `json:"result"`
}

// 添加验证者post参数
type AddReq struct {
	TotalBSCTMPubkeyPair []BTM `json:"bscTMPubkeyPairs"`
	PubkeyNum            int   `json:"pubkeyNum"`
	AccessToken          string
}

type TmTxErrResult struct {
	Jsonrpc string `json:"jsonrpc"`
	Id      int    `json:"id"`
	Error   struct {
		Code    int    `json:"code,omitempty"`
		Message string `json:"message,omitempty"`
		Data    string `json:"data,omitempty"`
	} `json:"error,omitempty"`
}
