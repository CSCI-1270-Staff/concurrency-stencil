package cursor

import (
	"dinodb/pkg/entry"
)

// Interface for a cursor that traverses a table.
type Cursor interface {
	Next() bool
	GetEntry() (entry.Entry, error)
}
