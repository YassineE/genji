// +build !wasm

package genji

import (
	"github.com/asdine/genji/engine"
	"github.com/asdine/genji/engine/badgerengine"
	"github.com/asdine/genji/engine/boltengine"
	"github.com/dgraph-io/badger/v2"
)

// Open creates a Genji database at the given path.
// If path is equal to ":memory:" it will open an in memory database,
// otherwise it will create an on-disk database using the BoltDB engine.
func Open(path string) (*DB, error) {
	var ng engine.Engine
	var err error

	switch path {
	case ":memory:":
		ng, err = badgerengine.NewEngine(badger.DefaultOptions("").WithInMemory(true).WithLogger(nil))
	default:
		ng, err = boltengine.NewEngine(path, 0660, nil)
	}
	if err != nil {
		return nil, err
	}

	return New(ng)
}
