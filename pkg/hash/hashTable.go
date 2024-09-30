package hash

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"

	"dinodb/pkg/entry"
	"dinodb/pkg/pager"
)

// A HashTable is a database index that uses extendible hashing for quick lookups
type HashTable struct {
	globalDepth int64        // The **global** depth of the Hash Table
	buckets     []int64      // Slice of bucket's page numbers. The indices (in binary) correspond to buckets' search keys in the HashTable
	pager       *pager.Pager // The pager associated with the Hash Table
}

// Returns a new HashTable.
func NewHashTable(pager *pager.Pager) (*HashTable, error) {
	depth := int64(2)
	buckets := make([]int64, powInt(2, depth))
	for i := range buckets {
		bucket, err := newHashBucket(pager, depth)
		if err != nil {
			return nil, err
		}
		buckets[i] = bucket.page.GetPageNum()
		pager.PutPage(bucket.page)
	}
	return &HashTable{globalDepth: depth, buckets: buckets, pager: pager}, nil
}

// Get depth.
func (table *HashTable) GetDepth() int64 {
	return table.globalDepth
}

// GetBuckets returns a slice containing the page numbers for all of this table's bucket.
func (table *HashTable) GetBuckets() []int64 {
	return table.buckets
}

// Get pager.
func (table *HashTable) GetPager() *pager.Pager {
	return table.pager
}

// Finds the entry with the given key.
func (table *HashTable) Find(key int64) (entry.Entry, error) {
	// Hash the key.
	hash := Hasher(key, table.globalDepth)
	if hash < 0 || int(hash) >= len(table.buckets) {
		return entry.Entry{}, errors.New("not found")
	}
	// Get the corresponding bucket.
	bucket, err := table.GetBucket(hash)
	if err != nil {
		return entry.Entry{}, err
	}
	defer table.pager.PutPage(bucket.page)

	// Find the entry.
	foundEntry, found := bucket.Find(key)
	if !found {
		return entry.Entry{}, errors.New("not found")
	}
	return foundEntry, nil
}

// ExtendTable increases the global depth of the table by 1.
func (table *HashTable) ExtendTable() {
	table.globalDepth = table.globalDepth + 1
	table.buckets = append(table.buckets, table.buckets...)
}

// Insert a key / value pair into the Hash Table.
func (table *HashTable) Insert(key int64, value int64) error {
	panic("Not yet implemented")
}

// Split the given bucket into two, extending the table if necessary.
//
// It is possible that after rehashing and redistributing, one of the buckets is empty
// and the other one still overflows, immediately requiring a second split.
// This may be a consequence of a bad hash function, but is a possible scenario
// that we should handle.
func (table *HashTable) split(bucket *HashBucket, hash int64) error {
	panic("Not yet implemented")
}

// Update the given key-value pair.
func (table *HashTable) Update(key int64, value int64) error {
	hash := Hasher(key, table.globalDepth)
	bucket, err := table.GetBucket(hash)
	if err != nil {
		return err
	}
	defer table.pager.PutPage(bucket.page)
	err2 := bucket.Update(key, value)
	return err2
}

// Delete the given key-value pair, does not coalesce.
func (table *HashTable) Delete(key int64) error {
	hash := Hasher(key, table.globalDepth)
	bucket, err := table.GetBucket(hash)
	if err != nil {
		return err
	}
	defer table.pager.PutPage(bucket.page)
	err2 := bucket.Delete(key)
	return err2
}

// Select all entries in this table.
func (table *HashTable) Select() ([]entry.Entry, error) {
	panic("Not yet implemented")
}

// Print writes a string representation of this entire table (including it's buckets) to the specified writer.
func (table *HashTable) Print(w io.Writer) {
	io.WriteString(w, "====\n")
	io.WriteString(w, fmt.Sprintf("global depth: %d\n", table.globalDepth))
	for i := range table.buckets {
		io.WriteString(w, fmt.Sprintf("====\nbucket %d\n", i))
		bucket, err := table.GetBucket(int64(i))
		if err != nil {
			continue
		}
		bucket.Print(w)
		table.pager.PutPage(bucket.page)
	}
	io.WriteString(w, "====\n")
}

// Print out a specific bucket.
func (table *HashTable) PrintPN(pn int, w io.Writer) {
	if int64(pn) >= table.pager.GetNumPages() {
		fmt.Println("out of bounds")
		return
	}
	bucket, err := table.GetBucket(int64(pn))
	if err != nil {
		return
	}
	bucket.Print(w)
	table.pager.PutPage(bucket.page)
}

/////////////////////////////////////////////////////////////////////////////
////////////////////////// HashTable Helper Functions ///////////////////////
/////////////////////////////////////////////////////////////////////////////

// Returns the bucket in the hash table using its page number, and increments the bucket ref count.
func (table *HashTable) GetBucketByPN(pn int64) (*HashBucket, error) {
	page, err := table.pager.GetPage(pn)
	if err != nil {
		return nil, err
	}
	return pageToBucket(page), nil
}

// Returns the bucket in the hash table, and increments the bucket ref count.
func (table *HashTable) GetBucket(hash int64) (*HashBucket, error) {
	pagenum := table.buckets[hash]
	bucket, err := table.GetBucketByPN(pagenum)
	if err != nil {
		return nil, err
	}
	return bucket, nil
}

// Read hash table in from memory.
func ReadHashTable(bucketPager *pager.Pager) (*HashTable, error) {
	backingFilename := bucketPager.GetFileName() + ".meta"
	indexPager, err := pager.New(backingFilename)
	if err != nil {
		return nil, err
	}
	metaPN := int64(0)
	metaPage, err := indexPager.GetPage(metaPN)
	if err != nil {
		return nil, err
	}
	// Read the gobal depth
	depth, _ := binary.Varint(metaPage.GetData()[:DEPTH_SIZE])
	bytesRead := DEPTH_SIZE
	// Read the bucket index
	pnSize := int64(binary.MaxVarintLen64)
	numHashes := powInt(2, depth)
	buckets := make([]int64, numHashes)
	for i := int64(0); i < numHashes; i++ {
		if bytesRead+pnSize > PAGESIZE {
			indexPager.PutPage(metaPage)
			metaPN++
			metaPage, err = indexPager.GetPage(metaPN)
			if err != nil {
				return nil, err
			}
			bytesRead = 0
		}
		pn, _ := binary.Varint(metaPage.GetData()[bytesRead : bytesRead+pnSize])
		bytesRead += pnSize
		buckets[i] = pn
	}
	indexPager.PutPage(metaPage)
	indexPager.Close()
	return &HashTable{globalDepth: depth, buckets: buckets, pager: bucketPager}, nil
}

// Write hash table out to memory.
func WriteHashTable(bucketPager *pager.Pager, table *HashTable) error {
	backingFilename := bucketPager.GetFileName() + ".meta"
	indexPager, err := pager.New(backingFilename)
	if err != nil {
		return err
	}
	metaPage, err := indexPager.GetNewPage()
	if err != nil {
		return err
	}
	metaPage.SetDirty(true)
	// Write global depth to meta file
	depthData := make([]byte, DEPTH_SIZE)
	binary.PutVarint(depthData, table.globalDepth)
	metaPage.Update(depthData, DEPTH_OFFSET, DEPTH_SIZE)
	bytesWritten := DEPTH_SIZE
	// Write bucket index to meta file
	pnSize := int64(binary.MaxVarintLen64)
	pnData := make([]byte, pnSize)
	for _, pn := range table.buckets {
		if bytesWritten+pnSize > PAGESIZE {
			indexPager.PutPage(metaPage)
			metaPage, err = indexPager.GetNewPage()
			if err != nil {
				return err
			}
			metaPage.SetDirty(true)
			bytesWritten = 0
		}
		binary.PutVarint(pnData, pn)
		metaPage.Update(pnData, bytesWritten, pnSize)
		bytesWritten += pnSize
	}
	indexPager.PutPage(metaPage)
	indexPager.Close()
	return bucketPager.Close()
}

// x^y
func powInt(x, y int64) int64 {
	return int64(math.Pow(float64(x), float64(y)))
}
