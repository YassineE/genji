package memory

import (
	"testing"

	"github.com/asdine/genji/field"
	"github.com/asdine/genji/record"
	"github.com/asdine/genji/table"
	"github.com/asdine/genji/table/tabletest"
	"github.com/stretchr/testify/require"
)

func TestMemoryEngineTable(t *testing.T) {
	tabletest.TestSuite(t, func() (table.Table, func()) {
		ng := NewEngine()
		tx, err := ng.Begin(true)
		require.NoError(t, err)

		tb, err := tx.CreateTable("test")
		require.NoError(t, err)

		return tb, func() {
			tx.Rollback()
			ng.Close()
		}
	})
}

func TestTable(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		ng := NewEngine()
		tx, err := ng.Begin(true)
		require.NoError(t, err)

		tr, err := tx.CreateTable("test")
		require.NoError(t, err)

		rec := record.FieldBuffer{
			field.NewString("name", "John"),
			field.NewInt64("age", 10),
		}
		rowid, err := tr.Insert(rec)
		require.NoError(t, err)

		resp, err := tr.Record(rowid)
		require.NoError(t, err)
		require.Equal(t, rec, resp)

		err = tx.Rollback()
		require.NoError(t, err)

		resp, err = tr.Record(rowid)
		require.Error(t, err)
		require.Nil(t, resp)

		require.Error(t, tx.Commit())

		require.NoError(t, ng.Close())
	})

	t.Run("iterate", func(t *testing.T) {
		ng := NewEngine()

		tx, err := ng.Begin(true)
		require.NoError(t, err)

		tr, err := tx.CreateTable("test")
		require.NoError(t, err)

		for i := int64(0); i < 3; i++ {
			rec := record.FieldBuffer{
				field.NewInt64("age", i),
			}
			_, err = tr.Insert(rec)
			require.NoError(t, err)
		}

		verifyContentFn := func(tab table.Table) {
			var i int64
			err := tab.Iterate(func(rec record.Record) bool {
				age, err := rec.Field("age")
				require.NoError(t, err)
				require.Equal(t, field.EncodeInt64(i), age.Data)
				i++
				return true
			})
			require.NoError(t, err)
			require.EqualValues(t, 3, i)
		}

		verifyContentFn(tr)

		err = tx.Commit()
		require.NoError(t, err)

		verifyContentFn(tr)
	})

	t.Run("undo table creation", func(t *testing.T) {
		tx, err := NewEngine().Begin(true)
		require.NoError(t, err)

		_, err = tx.CreateTable("test")
		require.NoError(t, err)
		err = tx.Rollback()
		require.NoError(t, err)
		_, err = tx.Table("test")
		require.Error(t, err)
	})
}
