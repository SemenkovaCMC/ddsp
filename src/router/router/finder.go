package router

import (
	"crypto/md5"
	"encoding/binary"
	"sort"

	"storage"
)

// Hasher is the common interface to compute hash for given k and node.
//
// Hasher это общий интерфейс для вычисления hash дла данных k и node.
type Hasher interface {
	Hash(k storage.RecordID, node storage.ServiceAddr) uint64
}

const poolSize = 4096

// MD5 implements Hasher interface computing hash base on md5 checksum.
//
// MD5 реализует интерфейс Hasher, вычисляя hash на основе контрольной суммы md5.
type MD5 struct {
	pool chan []byte
}

// NewMD5Hasher create new MD5.
func NewMD5Hasher() *MD5 {
	return &MD5{
		pool: make(chan []byte, poolSize),
	}
}

func (h *MD5) Hash(k storage.RecordID, node storage.ServiceAddr) uint64 {
	keySize := k.BinSize()
	size := keySize + node.BinSize()

	var buf []byte
	select {
	case buf = <-h.pool:
		if cap(buf) < size {
			buf = make([]byte, size)
		}
	default:
		buf = make([]byte, size)
	}

	buf = buf[:size]
	binary.LittleEndian.PutUint32(buf, uint32(k))
	copy(buf[keySize:], node)
	hash := md5.Sum(buf)

	select {
	case h.pool <- buf:
	default:
	}

	return binary.LittleEndian.Uint64(hash[:8])
}

// NodesFinder contains methods and options to find nodes where
// record with associated key shoud be stored.
//
// NodesFinder содержит методы и опции для нахождения узлов,
// на которых должна храниться запись с данным ключом.
type NodesFinder struct {
	h Hasher
}

// NewNodesFinder creates NodesFinder instance with given Hasher.
//
// NewNodesFinder создает NodesFinder с данным Hasher.
func NewNodesFinder(h Hasher) NodesFinder {
	return NodesFinder{h}
}

// NodesFind returns list of nodes where record with associated key k should be stored.
// Not more than storage.ReplciationFactor nodes is returned.
// Returned nodes are choosen from the provided slice of nodes.
//
// NodesFind возвращает список nodes, на которых должна храниться запись с ключом k.
// Возвращается не больше чем storage.ReplicationFactor nodes.
// Возвращаемые nodes выбираются из передаваемых nodes.
func (nf NodesFinder) NodesFind(k storage.RecordID, nodes []storage.ServiceAddr) []storage.ServiceAddr {

	hashedNodes := make([]struct {
		addr storage.ServiceAddr
		hash uint64
	}, 0, len(nodes))
	for _, node := range nodes {
		hashedNodes = append(hashedNodes, struct {
			addr storage.ServiceAddr
			hash uint64
		}{
			addr: node,
			hash: nf.h.Hash(k, node),
		})
	}

	sort.Slice(hashedNodes, func(i, j int) bool {
		if hashedNodes[i].hash == hashedNodes[j].hash {
			return hashedNodes[i].addr > hashedNodes[j].addr
		}
		return hashedNodes[i].hash > hashedNodes[j].hash
	})

	rf := storage.ReplicationFactor
	if len(hashedNodes) < rf {
		rf = len(hashedNodes)
	}

	selectedNodes := make([]storage.ServiceAddr, 0, rf)
	for i := 0; i < rf; i++ {
		selectedNodes = append(selectedNodes, hashedNodes[i].addr)
	}

	return selectedNodes
}
