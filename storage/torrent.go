package storage

// // FileDB. implement a db storage for save information of sending/downloading/downloaded files
// type TorrentDB struct {
// 	db   *LevelDBStore
// 	lock sync.RWMutex
// }

// type peerInfo struct {
// 	ID       common.PeerID
// 	Complete bool
// 	IP       [4]byte
// 	Port     uint16
// 	NodeAddr krpc.NodeAddr
// }

// type Torrent struct {
// 	Leechers int32
// 	Seeders  int32
// 	Peers    map[string]*peerInfo
// }

// func NewTorrentDB(db *LevelDBStore) *TorrentDB {
// 	return &TorrentDB{
// 		db: db,
// 	}
// }

// func (this *TorrentDB) Close() error {
// 	return this.db.Close()
// }

// // PutTorrent. put torrent info to db
// func (this *ChannelDB) AddPayment(hash string, torrent int32, amount uint64) error {
// 	key := []byte(fmt.Sprintf("payment:%d", paymentId))
// 	info := &Payment{
// 		WalletAddress: walletAddr,
// 		PaymentId:     paymentId,
// 		Amount:        amount,
// 	}
// 	buf, err := json.Marshal(info)
// 	if err != nil {
// 		return err
// 	}

// 	nodeAddr := krpc.NodeAddr{
// 		IP:   ar.IPAddress[:],
// 		Port: int(ar.Port),
// 	}
// 	log.Infof("ActionReg nodeAddr: %v, %d", nodeAddr.IP, nodeAddr.Port)
// 	nb, err := nodeAddr.MarshalBinary()
// 	log.Infof("ActionReg Wallet: %v, nb: %v", ar.Wallet.ToBase58(), nb)
// 	if err != nil {
// 		err = fmt.Errorf("nodeAddr marshal error")
// 		return err
// 	}
// 	err = storage.TDB.Put(ar.Wallet[:], nb)

// 	return this.db.Put(key, buf)
// }

// // GetPayment. get payment info from db
// func (this *ChannelDB) GetPayment(paymentId int32) (*Payment, error) {
// 	key := []byte(fmt.Sprintf("payment:%d", paymentId))
// 	value, err := this.db.Get(key)
// 	if err != nil {
// 		if err != leveldb.ErrNotFound {
// 			return nil, err
// 		}
// 	}
// 	if len(value) == 0 {
// 		return nil, nil
// 	}

// 	info := &Payment{}
// 	err = json.Unmarshal(value, info)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return info, nil
// }
