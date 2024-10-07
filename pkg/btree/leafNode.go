package btree

import (
	"dinodb/pkg/entry"
	"dinodb/pkg/pager"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
	"strconv"
)

// LeafNode represents a node at the bottom of a B+Tree that stores the actual
// key and value pairs that represent our data.
type LeafNode struct {
	NodeHeader           // Embeds all NodeHeader fields.
	rightSiblingPN int64 // The page number of the right sibling node.
}

// insert finds the appropriate place in the leaf node to insert a new key-value pair.
// If an entry already exists with the specified key, return an error. Returns a Split with
// relevant data to be used  by the caller if the insertion results in the node splitting.
//
// If the update flag is true, then insert will update the value of an existing key instead,
// returning an error if an existing entry to overwrite is not found.
func (node *LeafNode) insert(key int64, value int64, update bool) (Split, error) {
	panic("Not implemented yet")
}

// split is a helper function to split a leaf node, then propagate the split upwards.
func (node *LeafNode) split() (Split, error) {
	panic("Not implemented yet")
}

// delete removes a given key-value pair from the leaf node, if the given key exists.
func (node *LeafNode) delete(key int64) {
	// Find index of the specified key
	deletePos := node.search(key)
	if deletePos >= node.numKeys || node.getKeyAt(deletePos) != key {
		// Key was not found, so nothing to delete
		return
	}
	// Shift entries to the left, overwriting the key-value pair to be deleted
	for i := deletePos; i < node.numKeys-1; i++ {
		node.updateKeyAt(i, node.getKeyAt(i+1))
		node.updateValueAt(i, node.getValueAt(i+1))
	}
	node.updateNumKeys(node.numKeys - 1)
}

// get returns a boolean indicating whether the specified key was found,
// and if it was found, also returns the key's associated value.
func (node *LeafNode) get(key int64) (value int64, found bool) {
	// Find index of key
	index := node.search(key)
	if index >= node.numKeys || node.getKeyAt(index) != key {
		// Key was not found, so return false
		return 0, false
	}
	entry := node.getEntry(index)
	return entry.Value, true
}

/////////////////////////////////////////////////////////////////////////////
////////////////////////// Leaf Node  Helper Functions //////////////////////
/////////////////////////////////////////////////////////////////////////////

// search returns the first index where key >= given key.
// If no key satisfies this condition, returns numKeys.
func (node *LeafNode) search(key int64) int64 {
	// Binary search for the key.
	minIndex := sort.Search(
		int(node.numKeys),
		func(idx int) bool {
			return node.getKeyAt(int64(idx)) >= key
		},
	)
	return int64(minIndex)
}

// keyToNodeEntry is a helper function used to create cursors that point to the entry
// with the given key. Returns the node and index within that node where the entry is found.
func (node *LeafNode) keyToNodeEntry(key int64) (*LeafNode, int64, error) {
	return node, node.search(key), nil
}

// printNode pretty prints our leaf node.
func (node *LeafNode) printNode(w io.Writer, firstPrefix string, prefix string) {
	// Format header data.
	var nodeType string = "Leaf"
	var isRoot string
	if node.isRoot() {
		isRoot = " (root)"
	}
	numKeys := strconv.Itoa(int(node.numKeys))
	// Print header data.
	io.WriteString(w, fmt.Sprintf("%v[%v] %v%v size: %v\n",
		firstPrefix, node.page.GetPageNum(), nodeType, isRoot, numKeys))
	// Print entries.
	for Entrynum := int64(0); Entrynum < node.numKeys; Entrynum++ {
		entry := node.getEntry(Entrynum)
		io.WriteString(w, fmt.Sprintf("%v |--> (%v, %v)\n",
			prefix, entry.Key, entry.Value))
	}
	if node.rightSiblingPN > 0 {
		io.WriteString(w, fmt.Sprintf("%v |--+\n", prefix))
		io.WriteString(w, fmt.Sprintf("%v    | right sibling @ [%v]\n",
			prefix, node.rightSiblingPN))
		io.WriteString(w, fmt.Sprintf("%v    v\n", prefix))
	}
}

// pageToLeafNode returns the leaf node that is stored in the specified page.
func pageToLeafNode(page *pager.Page) *LeafNode {
	nodeHeader := pageToNodeHeader(page)
	rightSiblingPN, _ := binary.Varint(
		page.GetData()[RIGHT_SIBLING_PN_OFFSET : RIGHT_SIBLING_PN_OFFSET+RIGHT_SIBLING_PN_SIZE],
	)
	return &LeafNode{
		nodeHeader,
		rightSiblingPN,
	}
}

// createLeafNode creates and returns a new, empty leaf node.
// Nodes created with this function must use `PutPage()` accordingly after use.
func createLeafNode(pager *pager.Pager) (*LeafNode, error) {
	newPage, err := pager.GetNewPage()
	if err != nil {
		return &LeafNode{}, err
	}
	initPage(newPage, LEAF_NODE)
	return pageToLeafNode(newPage), nil
}

// getPage returns a pointer to the leaf node's page.
func (node *LeafNode) getPage() *pager.Page {
	return node.page
}

// getNodeType returns leafNode.
func (node *LeafNode) getNodeType() NodeType {
	return node.nodeType
}

// copy copies the metadata and data of the passed in LeafNode to this LeafNode.
func (node *LeafNode) copy(toCopy *LeafNode) {
	copy(node.page.GetData(), toCopy.page.GetData())
	node.updateNumKeys(toCopy.numKeys)
	node.setRightSibling(toCopy.rightSiblingPN)
}

// isRoot returns true if the current node is the root node.
func (node *LeafNode) isRoot() bool {
	return node.page.GetPageNum() == ROOT_PN
}

// setRightSibling sets the right sibling pagenumber field of the leaf node
// and updates the leaf node's page accordingly. Returns the old right sibling.
func (node *LeafNode) setRightSibling(siblingPN int64) int64 {
	// Retrieve the old sibling data
	oldSiblingPN := node.rightSiblingPN
	// Write the new sibling data to the page
	node.rightSiblingPN = siblingPN
	siblingData := make([]byte, RIGHT_SIBLING_PN_SIZE)
	binary.PutVarint(siblingData, node.rightSiblingPN)
	node.page.Update(
		siblingData,
		RIGHT_SIBLING_PN_OFFSET,
		RIGHT_SIBLING_PN_SIZE,
	)
	return oldSiblingPN
}

// entryPos returns the page offset to the entry at the given index.
func (node *LeafNode) entryPos(index int64) int64 {
	return LEAF_NODE_HEADER_SIZE + index*ENTRYSIZE
}

// modifyEntry updates the data stored in the entry at the given index.
func (node *LeafNode) modifyEntry(index int64, entry entry.Entry) {
	newdata := entry.Marshal()
	startPos := node.entryPos(index)
	node.page.Update(newdata, startPos, ENTRYSIZE)
}

// getEntry returns the entry stored in the entry at the given index.
func (node *LeafNode) getEntry(index int64) entry.Entry {
	startPos := node.entryPos(index)
	// Deserialize the entry.
	entry := entry.UnmarshalEntry(node.page.GetData()[startPos : startPos+ENTRYSIZE])
	return entry
}

// getKeyAt returns the key stored at the given index of the leaf node.
func (node *LeafNode) getKeyAt(index int64) int64 {
	return node.getEntry(index).Key
}

// updateKeyAt updates the key at the given index of the leaf node.
func (node *LeafNode) updateKeyAt(index int64, newKey int64) {
	existingVal := node.getValueAt(index)
	node.modifyEntry(index, entry.New(newKey, existingVal))
}

// getValueAt returns the value stored at the given index of the leaf node.
func (node *LeafNode) getValueAt(index int64) int64 {
	return node.getEntry(index).Value
}

// updateValueAt updates the value at the given index of the leaf node.
func (node *LeafNode) updateValueAt(index int64, newVal int64) {
	existingKey := node.getKeyAt(index)
	node.modifyEntry(index, entry.New(existingKey, newVal))
}

// updateNumKeys updates the numKeys field in the node struct and the page.
func (node *LeafNode) updateNumKeys(newNumKeys int64) {
	node.numKeys = newNumKeys
	// Write the new data to the page
	nKeysData := make([]byte, NUM_KEYS_SIZE)
	binary.PutVarint(nKeysData, newNumKeys)
	node.page.Update(nKeysData, NUM_KEYS_OFFSET, NUM_KEYS_SIZE)
}
