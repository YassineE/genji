package document

import (
	"encoding/csv"
	"encoding/json"
	"errors"
	"io"
)

// ErrStreamClosed is used to indicate that a stream must be closed.
var ErrStreamClosed = errors.New("stream closed")

// An Iterator can iterate over documents.
type Iterator interface {
	// Iterate goes through all the documents and calls the given function by passing each one of them.
	// If the given function returns an error, the iteration stops.
	Iterate(func(r Document) error) error
}

// NewIterator creates an iterator that iterates over documents.
func NewIterator(documents ...Document) Iterator {
	return documentsIterator(documents)
}

type documentsIterator []Document

func (rr documentsIterator) Iterate(fn func(r Document) error) error {
	var err error

	for _, r := range rr {
		err = fn(r)
		if err != nil {
			return err
		}
	}

	return nil
}

// Stream reads documents of an iterator one by one and passes them
// through a list of functions for transformation.
type Stream struct {
	it Iterator
	op StreamOperator
}

// NewStream creates a stream using the given iterator.
func NewStream(it Iterator) Stream {
	return Stream{it: it}
}

// Iterate calls the underlying iterator's iterate method.
// If this stream was created using the Pipe method, it will apply fn
// to any document passed by the underlying iterator.
// If fn returns a document, it will be passed to the next stream.
// If it returns a nil document, the document will be ignored.
// If it returns an error, the stream will be interrupted and that error will bubble up
// and returned by fn, unless that error is ErrStreamClosed, in which case
// the Iterate method will stop the iteration and return nil.
// It implements the Iterator interface.
func (s Stream) Iterate(fn func(r Document) error) error {
	if s.it == nil {
		return nil
	}

	if s.op == nil {
		return s.it.Iterate(fn)
	}

	opFn := s.op()

	err := s.it.Iterate(func(r Document) error {
		r, err := opFn(r)
		if err != nil {
			return err
		}

		if r == nil {
			return nil
		}

		return fn(r)
	})
	if err != ErrStreamClosed {
		return err
	}

	return nil
}

// Pipe creates a new Stream who can read its data from s and apply
// op to every document passed by its Iterate method.
func (s Stream) Pipe(op StreamOperator) Stream {
	return Stream{
		it: s,
		op: op,
	}
}

// Map applies fn to each received document and passes it to the next stream.
// If fn returns an error, the stream is interrupted.
func (s Stream) Map(fn func(r Document) (Document, error)) Stream {
	return s.Pipe(func() func(r Document) (Document, error) {
		return fn
	})
}

// Filter each received document using fn.
// If fn returns true, the document is kept, otherwise it is skipped.
// If fn returns an error, the stream is interrupted.
func (s Stream) Filter(fn func(r Document) (bool, error)) Stream {
	return s.Pipe(func() func(r Document) (Document, error) {
		return func(r Document) (Document, error) {
			ok, err := fn(r)
			if err != nil {
				return nil, err
			}

			if !ok {
				return nil, nil
			}

			return r, nil
		}
	})
}

// Limit interrupts the stream once the number of passed documents have reached n.
func (s Stream) Limit(n int) Stream {
	return s.Pipe(func() func(r Document) (Document, error) {
		var count int

		return func(r Document) (Document, error) {
			if count < n {
				count++
				return r, nil
			}

			return nil, ErrStreamClosed
		}
	})
}

// Offset ignores n documents then passes the subsequent ones to the stream.
func (s Stream) Offset(n int) Stream {
	return s.Pipe(func() func(r Document) (Document, error) {
		var skipped int

		return func(r Document) (Document, error) {
			if skipped < n {
				skipped++
				return nil, nil
			}

			return r, nil
		}
	})
}

// Append adds the given iterator to the stream.
func (s Stream) Append(it Iterator) Stream {
	if mr, ok := s.it.(multiIterator); ok {
		mr.iterators = append(mr.iterators, it)
		s.it = mr
	} else {
		s.it = multiIterator{
			iterators: []Iterator{s, it},
		}
	}

	return s
}

// Count counts all the documents from the stream.
func (s Stream) Count() (int, error) {
	counter := 0

	err := s.Iterate(func(r Document) error {
		counter++
		return nil
	})

	return counter, err
}

// First runs the stream, returns the first document found and closes the stream.
// If the stream is empty, all return values are nil.
func (s Stream) First() (r Document, err error) {
	err = s.Iterate(func(rec Document) error {
		r = rec
		return ErrStreamClosed
	})

	if err == ErrStreamClosed {
		err = nil
	}

	return
}

// An StreamOperator is used to modify a stream.
// If a stream operator returns a document, it will be passed to the next stream.
// If it returns a nil document, the document will be ignored.
// If it returns an error, the stream will be interrupted and that error will bubble up
// and returned by this function, unless that error is ErrStreamClosed, in which case
// the Iterate method will stop the iteration and return nil.
// Stream operators can be reused, and thus, any state or side effect should be kept within the operator closure
// unless the nature of the operator prevents that.
type StreamOperator func() func(r Document) (Document, error)

type multiIterator struct {
	iterators []Iterator
}

func (m multiIterator) Iterate(fn func(r Document) error) error {
	for _, it := range m.iterators {
		err := it.Iterate(fn)
		if err != nil {
			return err
		}
	}

	return nil
}

// IteratorToCSV encodes all the documents of an iterator to CSV.
func IteratorToCSV(w io.Writer, s Iterator) error {
	cw := csv.NewWriter(w)

	var line []string
	err := s.Iterate(func(r Document) error {
		line = line[:0]

		err := r.Iterate(func(f string, v Value) error {
			line = append(line, v.String())

			return nil
		})
		if err != nil {
			return err
		}

		return cw.Write(line)
	})
	if err != nil {
		return err
	}

	cw.Flush()
	return nil
}

// IteratorToJSON encodes all the documents of an iterator to JSON stream.
func IteratorToJSON(w io.Writer, s Iterator) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")

	return s.Iterate(func(r Document) error {
		return enc.Encode(jsonDocument{r})
	})
}