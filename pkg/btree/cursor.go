package btree

import (
	"errors"

	"dinodb/pkg/cursor"
	"dinodb/pkg/entry"
)

// BTreeCursor is a data structure that allows for easy iteration through
// the entries in a B+Tree's leaf nodes in order.
type BTreeCursor struct {
	index    *BTreeIndex // The B+Tree index that this cursor iterates through.
	curNode  *LeafNode   // Current leaf node we are pointing at
	curIndex int64       // The current index within curNode that we are pointing at.
}

// CursorAtStart returns a cursor pointing to the first entry of the B+Tree.
func (index *BTreeIndex) CursorAtStart() (cursor.Cursor, error) {
	// Get the root page.
	curPage, err := index.pager.GetPage(index.rootPN)
	if err != nil {
		return nil, err
	}
	defer index.pager.PutPage(curPage)
	curHeader := pageToNodeHeader(curPage)
	// Traverse down the leftmost children until we reach a leaf node.
	for curHeader.nodeType != LEAF_NODE {
		curNode := pageToInternalNode(curPage)
		leftmostPN := curNode.getPNAt(0)
		curPage, err = index.pager.GetPage(leftmostPN)
		if err != nil {
			return nil, err
		}
		defer index.pager.PutPage(curPage)
		curHeader = pageToNodeHeader(curPage)
	}
	// Set the cursor to point to the first entry in the leftmost leaf node.
	leftmostNode := pageToLeafNode(curPage)
	cursor := &BTreeCursor{index: index, curIndex: 0, curNode: leftmostNode}
	// Account for the edge case where the leftmostNode is empty
	// By adding a call to Next() here if the first node is empty,
	// we can guarantee that the cursor won't be stuck in an
	// empty node
	if cursor.curNode.numKeys == 0 {
		noEntries := cursor.Next()
		//if noEntries is true, then all our leaf nodes are empty
		if noEntries {
			return nil, errors.New("all leaf nodes are empty")
		}
	}
	return cursor, nil
}

// CursorAt returns a cursor pointing to the given key.
// If the key is not found, calls Next() to reach the next entry
// after the position of where key would be.
//
// Hint: use keyToNodeEntry
func (index *BTreeIndex) CursorAt(key int64) (cursor.Cursor, error) {
	// Get the root page.
	rootPage, err := index.pager.GetPage(index.rootPN)
	if err != nil {
		return nil, err
	}
	defer index.pager.PutPage(rootPage)
	rootNode := pageToNode(rootPage)
	// Find the leaf node and i that this key belongs to.
	leaf, i, err := rootNode.keyToNodeEntry(key)
	if err != nil {
		return nil, err
	}
	// Initialize cursor
	cursor := &BTreeCursor{index: index, curIndex: i, curNode: leaf}
	// If the cursor is not pointing at an entry, call Next()
	// This can happen if the entry associated 'key' was previously deleted
	// we can do this because CursorAt() is used only for SelectRange()
	if cursor.curIndex >= cursor.curNode.numKeys {
		cursor.Next()
	}
	return cursor, nil
}

// Next() moves the cursor ahead by one entry. Returns true at the end of the BTree.
func (cursor *BTreeCursor) Next() (atEnd bool) {
	// If the cursor is at the end of the node, go to the next node.
	if cursor.curIndex+1 >= cursor.curNode.numKeys {
		// Get the next node's page number.
		nextPN := cursor.curNode.rightSiblingPN
		if nextPN < 0 {
			return true
		}
		// Convert the page into a node.
		nextPage, err := cursor.index.pager.GetPage(nextPN)
		if err != nil {
			return true
		}
		defer cursor.index.pager.PutPage(nextPage)
		nextNode := pageToLeafNode(nextPage)
		// Reinitialize the cursor.
		cursor.curIndex = 0
		cursor.curNode = nextNode
		// If the next node is empty, step to the next node.
		// If no deletes are called, then this should never happen
		if nextNode.numKeys == 0 {
			return cursor.Next()
		}
		return false
	}
	// Else, just move the cursor forward.
	cursor.curIndex++
	return false
}

// GetEntry returns the entry currently pointed to by the cursor.
func (cursor *BTreeCursor) GetEntry() (entry.Entry, error) {
	// Check if we're retrieving a non-existent entry.
	if cursor.curIndex > cursor.curNode.numKeys {
		return entry.Entry{}, errors.New("getEntry: cursor is not pointing at a valid entry")
	}
	if cursor.curNode.numKeys == 0 {
		return entry.Entry{}, errors.New("getEntry: cursor is in an empty node :(")
	}
	entry := cursor.curNode.getEntry(cursor.curIndex)
	return entry, nil
}
