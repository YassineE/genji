package database

import (
	"github.com/asdine/genji/document"
	"github.com/asdine/genji/document/encoding"
	"github.com/asdine/genji/engine"
	"github.com/asdine/genji/index"
)

// TableConfig holds the configuration of a table
type TableConfig struct {
	PrimaryKeyName string
	PrimaryKeyType document.ValueType

	lastKey int64
}

type tableConfigStore struct {
	st engine.Store
}

func (t *tableConfigStore) Insert(tableName string, cfg TableConfig) error {
	key := []byte(tableName)
	_, err := t.st.Get(key)
	if err == nil {
		return ErrTableAlreadyExists
	}
	if err != engine.ErrKeyNotFound {
		return err
	}

	var fb document.FieldBuffer
	fb.Add("PrimaryKeyName", document.NewStringValue(cfg.PrimaryKeyName))
	fb.Add("PrimaryKeyType", document.NewUint8Value(uint8(cfg.PrimaryKeyType)))
	fb.Add("lastKey", document.NewInt64Value(cfg.lastKey))

	v, err := encoding.EncodeDocument(&fb)
	if err != nil {
		return err
	}

	return t.st.Put(key, v)
}

func (t *tableConfigStore) Replace(tableName string, cfg *TableConfig) error {
	key := []byte(tableName)
	_, err := t.st.Get(key)
	if err == engine.ErrKeyNotFound {
		return ErrTableNotFound
	}
	if err != nil {
		return err
	}

	var fb document.FieldBuffer
	fb.Add("PrimaryKeyName", document.NewStringValue(cfg.PrimaryKeyName))
	fb.Add("PrimaryKeyType", document.NewUint8Value(uint8(cfg.PrimaryKeyType)))
	fb.Add("lastKey", document.NewInt64Value(cfg.lastKey))

	v, err := encoding.EncodeDocument(&fb)
	if err != nil {
		return err
	}
	return t.st.Put(key, v)
}

func (t *tableConfigStore) Get(tableName string) (*TableConfig, error) {
	key := []byte(tableName)
	v, err := t.st.Get(key)
	if err == engine.ErrKeyNotFound {
		return nil, ErrTableNotFound
	}
	if err != nil {
		return nil, err
	}

	var cfg TableConfig

	r := encoding.EncodedDocument(v)

	f, err := r.GetByField("PrimaryKeyName")
	if err != nil {
		return nil, err
	}
	cfg.PrimaryKeyName, err = f.ConvertToString()
	if err != nil {
		return nil, err
	}
	f, err = r.GetByField("PrimaryKeyType")
	if err != nil {
		return nil, err
	}
	tp, err := f.ConvertToUint8()
	if err != nil {
		return nil, err
	}
	cfg.PrimaryKeyType = document.ValueType(tp)

	f, err = r.GetByField("lastKey")
	if err != nil {
		return nil, err
	}
	cfg.lastKey, err = f.ConvertToInt64()
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func (t *tableConfigStore) Delete(tableName string) error {
	key := []byte(tableName)
	err := t.st.Delete(key)
	if err == engine.ErrKeyNotFound {
		return ErrTableNotFound
	}
	return err
}

// Index of a table field. Contains information about
// the index configuration and provides methods to manipulate the index.
type Index struct {
	index.Index

	IndexName string
	TableName string
	FieldName string
	Unique    bool
}

type indexStore struct {
	st engine.Store
}

func (t *indexStore) Insert(cfg indexOptions) error {
	key := []byte(buildIndexName(cfg.IndexName))
	_, err := t.st.Get(key)
	if err == nil {
		return ErrIndexAlreadyExists
	}
	if err != engine.ErrKeyNotFound {
		return err
	}

	v, err := encoding.EncodeDocument(&cfg)
	if err != nil {
		return err
	}

	return t.st.Put(key, v)
}

func (t *indexStore) Get(indexName string) (*indexOptions, error) {
	key := []byte(buildIndexName(indexName))
	v, err := t.st.Get(key)
	if err == engine.ErrKeyNotFound {
		return nil, ErrIndexNotFound
	}
	if err != nil {
		return nil, err
	}

	var idxopts indexOptions
	err = idxopts.ScanDocument(encoding.EncodedDocument(v))
	if err != nil {
		return nil, err
	}

	return &idxopts, nil
}

func (t *indexStore) Delete(indexName string) error {
	key := []byte(buildIndexName(indexName))
	err := t.st.Delete(key)
	if err == engine.ErrKeyNotFound {
		return ErrIndexNotFound
	}
	return err
}
