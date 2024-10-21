package join

import (
	"context"
	"os"

	"dinodb/pkg/database"
	"dinodb/pkg/entry"
	"dinodb/pkg/hash"

	"golang.org/x/sync/errgroup"
)

const DEFAULT_FILTER_SIZE int64 = 1024

// Entry pair struct - output of a join.
type EntryPair struct {
	L entry.Entry
	R entry.Entry
}

// Int64 pair struct - to keep track of seen bucket pairs.
type pair struct {
	l int64
	r int64
}

// getTempDbFile Creates and returns the name of a temporary .db file used to back a pager.
func getTempDbFile() (string, error) {
	tmpfile, err := os.CreateTemp("", "*.db")
	if err != nil {
		return "", err
	}
	// close the temp file since it is open by default and we will open it again in Pager.New()
	_ = tmpfile.Close()
	return tmpfile.Name(), nil
}

// buildHashIndex constructs a temporary hash table for all the entries in the given sourceIndex.
// The useKey argument determines whether to use each entry's original key as the key
// in the temporary hash table (if false, we use the original value as the key).
func buildHashIndex(
	sourceIndex database.Index,
	useKey bool,
) (tempIndex *hash.HashIndex, dbName string, err error) {
	dbName, err = getTempDbFile()
	if err != nil {
		return nil, "", err
	}
	// Init the temporary hash table.
	tempIndex, err = hash.OpenTable(dbName)
	if err != nil {
		return nil, "", err
	}
	// Build the hash index.
	panic("Not yet implemented")
}

// sendResult attempts to send a single join result to the resultsChan channel as long as the errgroup hasn't been cancelled.
func sendResult(
	ctx context.Context,
	resultsChan chan EntryPair,
	result EntryPair,
) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case resultsChan <- result:
		return nil
	}
}

// probeBuckets sends the pair of entries in lBucket and rBucket that match on resultsChan.
//
// The joinOnLeftKey and joinOnRightKey arguments dictate whether we originally were matching on keys
// or values for each bucket. For example, with joinOnLeftKey = true and joinOnRightKey = false,
// we are finding the entries in lBucket whose keys match the value of entries in rBucket.
func probeBuckets(
	ctx context.Context,
	resultsChan chan EntryPair,
	lBucket *hash.HashBucket,
	rBucket *hash.HashBucket,
	joinOnLeftKey bool,
	joinOnRightKey bool,
) error {
	defer lBucket.GetPage().GetPager().PutPage(lBucket.GetPage())
	defer rBucket.GetPage().GetPager().PutPage(rBucket.GetPage())
	// Probe buckets.
	panic("Not yet implemented")
}

// Join leftIndex on rightIndex using the Grace Hash Join algorithm.
//
// The joinOnLeftKey and joinOnRightKey arguments dictate whether we are using keys or values
// for each bucket. For example, with joinOnLeftKey = true and joinOnRightKey = false,
// we are finding the entries in lBucket whose keys match the value of entries in rBucket.
func Join(
	ctx context.Context,
	leftIndex database.Index,
	rightIndex database.Index,
	joinOnLeftKey bool,
	joinOnRightKey bool,
) (resultsChan chan EntryPair, ctxt context.Context, group *errgroup.Group, cleanupCallback func(), err error) {
	// Create temporary hash tables for both tables
	leftHashIndex, leftDbName, err := buildHashIndex(leftIndex, joinOnLeftKey)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	rightHashIndex, rightDbName, err := buildHashIndex(rightIndex, joinOnRightKey)
	if err != nil {
		leftHashIndex.Close()
		os.Remove(leftDbName)
		os.Remove(leftDbName + ".meta")
		return nil, nil, nil, nil, err
	}
	cleanupCallback = func() {
		leftHashIndex.Close()
		rightHashIndex.Close()
		os.Remove(leftDbName)
		os.Remove(leftDbName + ".meta")
		os.Remove(rightDbName)
		os.Remove(rightDbName + ".meta")
	}
	// Make both hash indices the same global size.
	leftHashTable := leftHashIndex.GetTable()
	rightHashTable := rightHashIndex.GetTable()
	for leftHashTable.GetDepth() != rightHashTable.GetDepth() {
		if leftHashTable.GetDepth() < rightHashTable.GetDepth() {
			// Split the left table
			leftHashTable.ExtendTable()
		} else {
			// Split the right table
			rightHashTable.ExtendTable()
		}
	}
	// Probe phase: match buckets to buckets and emit entries that match.
	group, ctx = errgroup.WithContext(ctx)
	resultsChan = make(chan EntryPair, 1024)
	// Iterate through hash buckets, keeping track of pairs we've seen before.
	leftBuckets := leftHashTable.GetBuckets()
	rightBuckets := rightHashTable.GetBuckets()
	seenList := make(map[pair]bool)
	for i, lBucketPN := range leftBuckets {
		rBucketPN := rightBuckets[i]
		bucketPair := pair{l: lBucketPN, r: rBucketPN}
		if _, seen := seenList[bucketPair]; seen {
			continue
		}
		seenList[bucketPair] = true

		lBucket, err := leftHashTable.GetBucketByPN(lBucketPN)
		if err != nil {
			return nil, nil, nil, cleanupCallback, err
		}
		rBucket, err := rightHashTable.GetBucketByPN(rBucketPN)
		if err != nil {
			lBucket.GetPage().GetPager().PutPage(lBucket.GetPage())
			return nil, nil, nil, cleanupCallback, err
		}
		group.Go(func() error {
			return probeBuckets(ctx, resultsChan, lBucket, rBucket, joinOnLeftKey, joinOnRightKey)
		})
	}
	return resultsChan, ctx, group, cleanupCallback, nil
}
