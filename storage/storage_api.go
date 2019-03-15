/**
 * Description:
 * Author: LiYong Zhang
 * Create: 2019-03-14 
*/
package storage

import "github.com/oniio/oniChain/common/log"

type Storage struct {
	Ls *LevelDBStore
}

func NewStorage(path string) *Storage{
	nls, err := NewLevelDBStore(path)
	if err != nil {
		log.Errorf("init torrent cache err:%s\n", err)
		return nil
	}
	return &Storage{Ls:nls}
}

//Put a key-value pair to leveldb
func (self *Storage) Put(key []byte, value []byte) error {
	return self.Ls.Put(key, value)
}

//Get the value of a key from leveldb
func (self *Storage) Get(key []byte) ([]byte, error) {
	return self.Ls.Get(key)

}

//Has return whether the key is exist in leveldb
func (self *Storage) Has(key []byte) (bool, error) {
	return self.Ls.Has(key)
}

//Delete the the in leveldb
func (self *Storage) Delete(key []byte) error {
	return self.Ls.Delete(key)
}
