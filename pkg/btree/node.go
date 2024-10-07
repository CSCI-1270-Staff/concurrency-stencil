package btree

import (
	"encoding/binary"
	"io"

	"dinodb/pkg/pager"
)

/////////////////////////////////////////////////////////////////////////////
///////////////////////// Structs and interfaces ////////////////////////////
/////////////////////////////////////////////////////////////////////////////

// Split is a supporting data structure to propagate information
// needed to implement splits up our B+tree after inserts.
type Split struct {
	isSplit bool  // A flag that's set to true if a split occurs.
	key     int64 // The median key that is being pushed up.
	leftPN  int64 // The pagenumber for the left node.
	rightPN int64 // The pagenumber for the right node.
}

// Node defines a common interface for leaf and internal nodes.
type Node interface {
	// insert traverses down the B+Tree and inserts the specified
	// key-value pair into a leaf node. Cannot insert a duplicate key.
	// Returns a Split with relevant data to be used by the caller
	// if the insertion results in the node splitting.
	//
	// If the update flag is true, then insert will perform an update instead,
	// returning an error if an existing entry to overwrite is not found.
	insert(key int64, value int64, update bool) (Split, error)

	// delete traverses down the B+Tree and removes the entry with the given key
	// from the leaf nodes if it exists.
	// Note that delete does not implement merging of node (see handout for more details).
	delete(key int64)

	// get tries to find the value associated with the given key in the B+Tree,
	// traversing down to the leaf nodes. It returns a boolean indicating whether
	// the key was found in the node and the associated value if found.
	get(key int64) (value int64, found bool)

	// Helper methods added for convenience
	search(searchKey int64) int64
	// keyToNodeEntry is a helper function used to create cursors that point to the entry
	// with the given key. Returns the node and index within that node where the entry is found.
	keyToNodeEntry(key int64) (node *LeafNode, index int64, err error)
	// printNode writes a string representation of the node to the specified
	printNode(io.Writer, string, string)
	// getPage returns the node's underlying page where it's data is stored.
	getPage() *pager.Page
	getNodeType() NodeType
}

// NodeType identifies if a node is a leaf node or an internal node.
type NodeType bool

const (
	INTERNAL_NODE NodeType = false
	LEAF_NODE     NodeType = true
)

// NodeHeaders contain metadata common to all types of nodes
type NodeHeader struct {
	nodeType NodeType    // The type of the node (either leaf or internal).
	numKeys  int64       // The number of keys currently stored in the node.
	page     *pager.Page // The page that holds the node's data
}

/////////////////////////////////////////////////////////////////////////////
//////////////////////// Generic Helper Functions ///////////////////////////
/////////////////////////////////////////////////////////////////////////////

// initPage resets the page's data then sets the nodeType bit.
func initPage(page *pager.Page, nodeType NodeType) {
	page.SetDirty(true)
	copy(page.GetData(), make([]byte, pager.Pagesize))
	if nodeType == LEAF_NODE {
		(page.GetData())[NODETYPE_OFFSET] = 1 // Set the nodeType bit
	}
}

// pageToNode returns the node corresponding to the given page.
func pageToNode(page *pager.Page) Node {
	nodeHeader := pageToNodeHeader(page)
	if nodeHeader.nodeType == LEAF_NODE {
		return pageToLeafNode(page)
	}
	return pageToInternalNode(page)
}

// pageToNodeHeader returns node header data from the given page.
func pageToNodeHeader(page *pager.Page) NodeHeader {
	var nodeType NodeType
	if page.GetData()[NODETYPE_OFFSET] == 0 {
		nodeType = INTERNAL_NODE
	} else {
		nodeType = LEAF_NODE
	}
	numKeys, _ := binary.Varint(
		page.GetData()[NUM_KEYS_OFFSET : NUM_KEYS_OFFSET+NUM_KEYS_SIZE],
	)
	return NodeHeader{
		nodeType: nodeType,
		numKeys:  numKeys,
		page:     page,
	}
}
