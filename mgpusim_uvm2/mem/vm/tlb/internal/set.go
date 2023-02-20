// Package internal provides the definition required for defining TLB.
package internal

import (
	"fmt"

	"github.com/google/btree"
	"gitlab.com/akita/mem/v2/vm"
	"gitlab.com/akita/util/v2/ca"
)

// A Set holds a certain number of pages.
type Set interface {
	Lookup(pid ca.PID, vAddr uint64) (wayID int, page vm.Page, found bool)
	Update(wayID int, page vm.Page)
	Evict() (wayID int, ok bool)
	Visit(wayID int)
	DeletePage(wayID int, page vm.Page)
}

// NewSet creates a new TLB set.
func NewSet(numWays int) Set {
	s := &setImpl{}
	s.blocks = make([]*block, numWays)
	s.visitTree = btree.New(2)
	s.vAddrWayIDMap = make(map[string]int)
	for i := range s.blocks {
		b := &block{}
		s.blocks[i] = b
		b.wayID = i
		s.Visit(i)
	}
	return s
}

type block struct {
	page      vm.Page
	wayID     int
	lastVisit uint64
}

func (b *block) Less(anotherBlock btree.Item) bool {
	return b.lastVisit < anotherBlock.(*block).lastVisit
}

type setImpl struct {
	blocks        []*block
	vAddrWayIDMap map[string]int
	visitTree     *btree.BTree
	visitCount    uint64
}

func (s *setImpl) keyString(pid ca.PID, vAddr uint64) string {
	return fmt.Sprintf("%d%016x", pid, vAddr)
}

func (s *setImpl) Lookup(pid ca.PID, vAddr uint64) (
	wayID int,
	page vm.Page,
	found bool,
) {
	key := s.keyString(pid, vAddr)
	wayID, ok := s.vAddrWayIDMap[key]
	if !ok {
		return 0, vm.Page{}, false
	}

	block := s.blocks[wayID]

	return block.wayID, block.page, true
}

func (s *setImpl) Update(wayID int, page vm.Page) {
	block := s.blocks[wayID]
	key := s.keyString(block.page.PID, block.page.VAddr)
	delete(s.vAddrWayIDMap, key)

	block.page = page
	key = s.keyString(page.PID, page.VAddr)
	s.vAddrWayIDMap[key] = wayID
}

func (s *setImpl) Evict() (wayID int, ok bool) {
	if s.hasNothingToEvict() {
		return 0, false
	}

	wayID = s.visitTree.DeleteMin().(*block).wayID
	return wayID, true
}

func (s *setImpl) Visit(wayID int) {
	block := s.blocks[wayID]
	s.visitTree.Delete(block)

	s.visitCount++
	block.lastVisit = s.visitCount
	s.visitTree.ReplaceOrInsert(block)
}

func (s *setImpl) hasNothingToEvict() bool {
	return s.visitTree.Len() == 0
}


// delete Page, delete this page's block
func (s *setImpl) DeletePage(wayID int, page vm.Page) {
	//fmt.Println("------\n",s.blocks)
	block := s.blocks[wayID]
	key := s.keyString(block.page.PID, block.page.VAddr)

	s.blocks = append(s.blocks[:wayID], s.blocks[wayID+1:]...)
	//s.visitTree.Delete(block)
	//s.visitCount--
	delete(s.vAddrWayIDMap, key)

	//fmt.Println(s.blocks)
}

