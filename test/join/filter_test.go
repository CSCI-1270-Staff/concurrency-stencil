package join_test

import (
	"dinodb/pkg/join"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	t.Run("Create", testFilterCreate)
	t.Run("InsertAndCheckSmall", testFilterInsertAndCheckSmall)
	t.Run("SizeSmall", testFilterSizeSmall)
}

func testFilterCreate(t *testing.T) {
	filter := join.CreateFilter(16)
	for i := 0; i < 1000; i++ {
		if filter.Contains(int64(i)) {
			t.Error("new filter should be empty")
		}
	}
}

func testFilterInsertAndCheckSmall(t *testing.T) {
	filter := join.CreateFilter(16)
	for i := 0; i < 10; i++ {
		filter.Insert(int64(i))
		if !filter.Contains(int64(i)) {
			t.Errorf("inserted value %d but not found", i)
		}
	}
}

func testFilterSizeSmall(t *testing.T) {
	filter := join.CreateFilter(1)
	filter.Insert(int64(0))
	for i := 0; i < 100; i++ {
		if !filter.Contains(int64(i)) {
			t.Errorf("should have 'found' value %d", i)
		}
	}
}
