// Package pager implements the page and pager abstractions used for efficient io operations in our database
package pager

import (
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	"dinodb/pkg/config"
	"dinodb/pkg/list"

	"github.com/ncw/directio"
)

// Pagesize is the size of an individual page (ie the maximum number of bytes that the page can hold) - defaults to 4kb.
const Pagesize int64 = directio.BlockSize

// Error for when there are no free/unpinned pages to be used
var ErrRanOutOfPages = errors.New("no available pages")

// Pager is a data structure that manages pages of data stored in a file.
type Pager struct {
	file         *os.File   // File descriptor for the file that backs this pager on disk.
	numPages     int64      // The number of pages that this page has access to (both on disk and in memory).
	freeList     *list.List // A list of pre-allocated (but unused) pages.
	unpinnedList *list.List // The list of pages in memory that have yet to be evicted, but are not currently in use.
	pinnedList   *list.List // The list of in-memory pages currently being used by the database.
	// The page table, which maps pagenums to their corresponding pages (stored in a link belonging to the list the page is in).
	pageTable map[int64]*list.Link
	ptMtx     sync.Mutex // Mutex for protecting the Page table for concurrent use.
}

// New constructs a new Pager, backing it with a database file at the specified filePath.
// See [*Pager.Open] for more details on backing the Pager with database files.
func New(filePath string) (pager *Pager, err error) {
	pager = &Pager{}
	pager.pageTable = make(map[int64]*list.Link)
	pager.freeList = list.NewList()
	pager.unpinnedList = list.NewList()
	pager.pinnedList = list.NewList()
	frames := directio.AlignedBlock(int(Pagesize * config.MaxPagesInBuffer))
	for i := 0; i < config.MaxPagesInBuffer; i++ {
		frame := frames[i*int(Pagesize) : (i+1)*int(Pagesize)]
		page := Page{
			pager:   pager,
			pagenum: NoPage,
			dirty:   false,
			data:    frame,
		}
		pager.freeList.PushTail(&page)
	}

	err = pager.Open(filePath)
	if err != nil {
		pager = nil
	}
	return
}

// GetFileName returns the file name/path used to open the pager's backing file.
func (pager *Pager) GetFileName() (filename string) {
	return pager.file.Name()
}

// GetNumPages returns the number of pages.
func (pager *Pager) GetNumPages() (numPages int64) {
	return pager.numPages
}

// GetFreePN returns the next available page number.
func (pager *Pager) GetFreePN() (nextPN int64) {
	// Assign the first page number beyond the end of the file.
	return pager.numPages
}

// Open (re-)initializes our pager with a database file at the specified filePath.
//
// If the database file didn't exist previously, it is created.
// If the database file does exist but it can't be opened or
// it's contents are not properly aligned to PAGESIZE, returns an error.
// The Pager should not be used if an error is returned.
func (pager *Pager) Open(filePath string) (err error) {
	// Create the necessary prerequisite directories.
	if idx := strings.LastIndex(filePath, "/"); idx != -1 {
		err = os.MkdirAll(filePath[:idx], 0775)
		if err != nil {
			return err
		}
	}
	// Open or create the db file.
	pager.file, err = directio.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	// Get info about the size of the pager.
	var info os.FileInfo
	var len int64
	if info, err = pager.file.Stat(); err == nil {
		len = info.Size()
		if len%Pagesize != 0 {
			return errors.New("DB file has been corrupted")
		}
	}
	// Set the number of pages and hand off initialization to someone else.
	pager.numPages = len / Pagesize
	return nil
}

// Close signals our pager to flush all dirty pages to disk
// and close its backing file.
func (pager *Pager) Close() error {
	// Prevent new data from being paged in.
	pager.ptMtx.Lock()
	defer pager.ptMtx.Unlock()
	// Check that no pages are in the pinned list
	curLink := pager.pinnedList.PeekHead()
	if curLink != nil {
		return errors.New("pages are still pinned on close")
	}
	// Cleanup.
	pager.FlushAllPages()
	return pager.file.Close()
}

// FillPageFromDisk populate a page's data field from the data currently on disk.
// Returns an error if there was an io problem reading from disk.
func (pager *Pager) FillPageFromDisk(page *Page) error {
	if _, err := pager.file.Seek(page.pagenum*Pagesize, 0); err != nil {
		return err
	}
	if _, err := pager.file.Read(page.data); err != nil && err != io.EOF {
		return err
	}
	return nil
}

// newPage returns a currently unused Page from the free or unpinned list,
// or an ErrRanOutOfPages if there are no unused pages available.
// The ptMtx should be locked on entry.
func (pager *Pager) newPage(pagenum int64) (newPage *Page, err error) {
	panic("Not implemented yet")
}

// GetNewPage returns a new Page with the next available pagenum
func (pager *Pager) GetNewPage() (page *Page, err error) {
	panic("Not implemented yet")
}

// GetPage returns an existing Page corresponding to the given pagenum.
func (pager *Pager) GetPage(pagenum int64) (page *Page, err error) {
	panic("Not implemented yet")
}

// PutPage releases a reference to a page.
func (pager *Pager) PutPage(page *Page) (err error) {
	pager.ptMtx.Lock()
	defer pager.ptMtx.Unlock()
	// Decrement pinCount
	ret := page.Put()
	// Check if we can unpin this page; if so, move from pinned to unpinned list.
	if ret == 0 {
		link := pager.pageTable[page.pagenum]
		link.PopSelf()
		newLink := pager.unpinnedList.PushTail(page)
		pager.pageTable[page.pagenum] = newLink
	}
	if ret < 0 {
		return errors.New("pinCount for page is < 0")
	}
	return nil
}

// FlushPage flushes a particular page's data to disk if it is dirty.
func (pager *Pager) FlushPage(page *Page) {
	panic("Not implemented yet")
}

// FlushAllPages flushes all dirty pages to disk.
func (pager *Pager) FlushAllPages() {
	panic("Not implemented yet")
}
