package mypersistmap

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
)

//PersistMap PersistMap
type PersistMap struct {
	db     *leveldb.DB
	mapVal map[int64]int64
	sync.RWMutex
}

//Get Get
//key 不能为0
func (v *PersistMap) Get(key int64) int64 {
	v.RLock()
	rltv, _ := v.mapVal[key]
	v.RUnlock()
	return rltv
}

//Set Set
func (v *PersistMap) Set(key int64, value int64) {
	v.db.Put(Int64ToBytes(key), Int64ToBytes(value), nil)

	v.Lock()
	v.mapVal[key] = value
	v.Unlock()
}

//Len Len
func (v *PersistMap) Len() int {
	v.RLock()
	length := len(v.mapVal)
	v.RUnlock()
	return length
}

// //InitPersistMap InitPersistMap
// func (v *PersistMap) InitPersistMap(path string) error {
// 	v.mapVal = make(map[int64]int64)

// 	var err error
// 	v.db, err = leveldb.OpenFile(path, nil)
// 	if v.db == nil {
// 		fmt.Println("err:", err)
// 		return err
// 	}

// 	iter := v.db.NewIterator(nil, nil)
// 	for iter.Next() {
// 		key := int64(binary.BigEndian.Uint64(iter.Key()))
// 		value := int64(binary.BigEndian.Uint64(iter.Value()))

// 		v.Set(key, value)
// 	}
// 	iter.Release()
// 	err = iter.Error()

// 	return err
// }

//Int64ToBytes Int64ToBytes
func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

//BytesToInt64 BytesToInt64
func BytesToInt64(buf []byte) int64 {
	return int64(binary.BigEndian.Uint64(buf))
}

//NewPersistMap NewPersistMap
func NewPersistMap(path string) *PersistMap {

	var v PersistMap
	v.mapVal = make(map[int64]int64)

	var err error
	v.db, err = leveldb.OpenFile(path, nil)
	if v.db == nil {
		fmt.Println("err:", err)
		return nil
	}

	iter := v.db.NewIterator(nil, nil)
	for iter.Next() {
		key := int64(binary.BigEndian.Uint64(iter.Key()))
		value := int64(binary.BigEndian.Uint64(iter.Value()))

		v.Set(key, value)
	}
	iter.Release()
	err = iter.Error()

	fmt.Println("NewPersistMap ", path, "len map :", v.Len())

	return &v
}
