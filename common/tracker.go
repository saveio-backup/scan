package common

import (
	"encoding/hex"
)

type MetaInfoHash [46]byte // MetaInfo hash
type PeerID [20]byte       // Peer's ID

// tracker
const (
	MAX_TRACKER_TORRENT_SIZE      = 10000     // Max torrent size of tracker
	MAX_TRACKER_TORRENT_KEEP_TIME = 24 * 3600 // Max torrent keep time in second
)

func MetaInfoHashToString(hash MetaInfoHash) string {
	return hex.EncodeToString(hash[:])
}

func MetaInfoHashFromString(str string) (MetaInfoHash, error) {
	v, err := hex.DecodeString(str)
	if err != nil {
		return MetaInfoHash{}, err
	}
	var dst MetaInfoHash
	copy(dst[:], v[:])
	return dst, nil
}
