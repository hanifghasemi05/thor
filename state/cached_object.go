package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/vechain/thor/cry"
	"github.com/vechain/thor/kv"
)

// cachedObject to cache code and storage of an account.
type cachedObject struct {
	kv   kv.GetPutter // needed for creating storage trie
	data Account

	cache struct {
		code        []byte
		storageTrie trieReader
		storage     map[cry.Hash]cry.Hash
	}
}

func newCachedObject(kv kv.GetPutter, data Account) *cachedObject {
	return &cachedObject{kv: kv, data: data}
}

// GetStorage returns storage value for given key.
func (co *cachedObject) GetStorage(key cry.Hash) (cry.Hash, error) {
	cache := &co.cache
	// retrive from storage cache
	if cache.storage == nil {
		cache.storage = make(map[cry.Hash]cry.Hash)
	} else {
		if v, ok := cache.storage[key]; ok {
			return v, nil
		}
	}
	// not found in cache

	// create storage trie if not yet created
	if cache.storageTrie == nil {
		root := common.BytesToHash(co.data.StorageRoot)
		trie, err := trie.NewSecure(root, co.kv, 0)
		if err != nil {
			return cry.Hash{}, err
		}
		cache.storageTrie = trie
	}

	// load from trie
	v, err := loadStorage(cache.storageTrie, key)
	if err != nil {
		return cry.Hash{}, err
	}
	// put into cache
	cache.storage[key] = v
	return v, nil
}

// GetCode returns the code of the account.
func (co *cachedObject) GetCode() ([]byte, error) {
	cache := &co.cache

	if len(cache.code) > 0 {
		return cache.code, nil
	}

	if len(co.data.CodeHash) > 0 {
		// do have code
		code, err := co.kv.Get(co.data.CodeHash)
		if err != nil {
			return nil, err
		}
		cache.code = code
		return code, nil
	}
	return nil, nil
}