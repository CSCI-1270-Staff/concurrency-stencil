package utils

// Interface for an entry in a table.
type Entry interface {
	GetKey() int64
	GetValue() int64
	SetKey(key int64)
	SetValue(value int64)
	Marshal() []byte
}
