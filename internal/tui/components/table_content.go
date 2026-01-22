package components

import (
	"container/list"
	"sync"

	"github.com/rivo/tview"

	"github.com/android-lewis/dbsmith/internal/tui/constants"
)

type CellKey struct {
	Row int
	Col int
}

type cacheEntry struct {
	key  CellKey
	cell *tview.TableCell
}

type LRUCellCache struct {
	maxSize int
	cache   map[CellKey]*list.Element
	lruList *list.List
	mu      sync.RWMutex
}

func NewLRUCellCache(maxSize int) *LRUCellCache {
	if maxSize <= 0 {
		maxSize = constants.DefaultCellCacheSize
	}
	return &LRUCellCache{
		maxSize: maxSize,
		cache:   make(map[CellKey]*list.Element),
		lruList: list.New(),
	}
}

func (c *LRUCellCache) Get(row, col int) *tview.TableCell {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := CellKey{Row: row, Col: col}
	if elem, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(elem)
		return elem.Value.(*cacheEntry).cell
	}
	return nil
}

func (c *LRUCellCache) Set(row, col int, cell *tview.TableCell) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := CellKey{Row: row, Col: col}

	if elem, ok := c.cache[key]; ok {
		c.lruList.MoveToFront(elem)
		elem.Value.(*cacheEntry).cell = cell
		return
	}

	if c.lruList.Len() >= c.maxSize {
		c.evictOldest()
	}

	entry := &cacheEntry{key: key, cell: cell}
	elem := c.lruList.PushFront(entry)
	c.cache[key] = elem
}

func (c *LRUCellCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[CellKey]*list.Element)
	c.lruList.Init()
}

func (c *LRUCellCache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.lruList.Len()
}

func (c *LRUCellCache) evictOldest() {
	if elem := c.lruList.Back(); elem != nil {
		entry := elem.Value.(*cacheEntry)
		delete(c.cache, entry.key)
		c.lruList.Remove(elem)
	}
}
