package tlb

import (
	"log"
	"reflect"
	"strings"
	"fmt"
	"gitlab.com/akita/akita/v3/sim"
	"gitlab.com/akita/akita/v3/tracing"
	"gitlab.com/akita/mem/v3/vm"
	"gitlab.com/akita/mem/v3/vm/tlb/internal"

)


type Table struct{
	pid    vm.PID
	vaddr         uint64
	store[9]      int
	count         int
	access[9]     int
}

type iommu struct{
	pid    vm.PID
	addr   uint64
}

type l2 struct{
	pid    vm.PID
	addr   uint64
}


var 
(
	GPU_NUMBER 				int=4
	setpagesize			uint64 = 4096
	i,j,k,p,h 				int
	table               []Table
	tmp                 []Table
	store[9]            int
	access[9]				int
	existloc            int = -1
	gid			  		int
	pagenumber          int
	sharePercentage[9]  int
	accesstime[9]       int
	singleaccess		int
	cycle 				int = 0
	accesscounter 		int = 1

	mmuset              int = 64        
	mmuway              int = 32
	mmutlb              [64][32]iommu
	mmusetID            int
	mmutlblength        int
	mmutlblocation      int
	mmufoundflag        int
	tablefoundflag      int
	mmuhit[5]           int
	mmumiss[5]          int
	mmucount[5]         int

	lastvisit			[]Last
	remoteaccess[9]		int
	localaccess[9]		int
	remoteaccesstimes[9] int

	numGPUs				int = 9

	L2miss				int = 0
	L2hit				int = 0
	L2mshr				int = 0
	unnecessary[9]			int 
	localaccessGPU[9]	int

	//l1 and l2 tlb
	l2set               int=32
	l2way               int=16
	l2tlb               [8][32][16]l2 
	// l2set  				int=64
	// l2way 	 			int=32
	// l2tlb[4][64][32]    uint64

	
	l2tlblength         int

	

	l2tlblocation       int
	l2tlbtmp            uint64
	l2setID             int
	l2hit               int=0
	l1hit               int=0
	l2miss              int=0
	l1miss              int=0
	l2foundflag         int
	pagesize            int= 1<<16
	evictl2tlb          uint64

	// fully associate L1
	//l1 1个cu 32个entry
	// l1 tlb 4gpu 64cu 32entry
	//l1 三个类型
	l1tlb[8][64][32]    uint64 //4是gpu个数， 64是CU个数，16是每个L1 TLB 大小

	//to check
	total_gmmu_hit 		int=0

)

type Last struct{
	pid    vm.PID
	vaddr       uint64
	visitid  	uint64 //这个page在哪
	remoteaccesstime[9]   int //gpu访问记录
	
}

// A TLB is a cache that maintains some page information.
type TLB struct {
	*sim.TickingComponent

	topPort     sim.Port
	bottomPort  sim.Port
	controlPort sim.Port

	LowModule sim.Port

	numSets        int
	numWays        int
	pageSize       uint64
	numReqPerCycle int

	Sets []internal.Set 

	mshr                mshr
	respondingMSHREntry *mshrEntry

	name           string
	isPaused bool
	isL2            bool
}



// Reset sets all the entries int he TLB to be invalid
func (tlb *TLB) reset() {
	tlb.Sets = make([]internal.Set, tlb.numSets)
	for i := 0; i < tlb.numSets; i++ {
		set := internal.NewSet(tlb.numWays)
		tlb.Sets[i] = set
	}

}

// Tick defines how TLB update states at each cycle
func (tlb *TLB) Tick(now sim.VTimeInSec) bool {

	madeProgress := false

	madeProgress = tlb.performCtrlReq(now) || madeProgress

	if !tlb.isPaused {
		for i := 0; i < tlb.numReqPerCycle; i++ {
			madeProgress = tlb.respondMSHREntry(now) || madeProgress
		}

		for i := 0; i < tlb.numReqPerCycle; i++ {
			madeProgress = tlb.lookup(now) || madeProgress
		}

		for i := 0; i < tlb.numReqPerCycle; i++ {
			madeProgress = tlb.parseBottom(now) || madeProgress
		}
		cycle ++
	}

	return madeProgress
}


func (tlb *TLB) respondMSHREntry(now sim.VTimeInSec) bool {
	if tlb.respondingMSHREntry == nil {
		return false
	}

	mshrEntry := tlb.respondingMSHREntry
	page := mshrEntry.page
	req := mshrEntry.Requests[0]
	rspToTop := vm.TranslationRspBuilder{}.
		WithSendTime(now).
		WithSrc(tlb.topPort).
		WithDst(req.Src).
		WithRspTo(req.ID).
		WithPage(page).
		Build()
	err := tlb.topPort.Send(rspToTop)
	if err != nil {
		return false
	}

	mshrEntry.Requests = mshrEntry.Requests[1:]
	if len(mshrEntry.Requests) == 0 {
		tlb.respondingMSHREntry = nil
	}

	tracing.TraceReqComplete(req, tlb)
	return true
}


// 对应的vaddr在lastvisit的index
func (tlb *TLB) Lastvisitlookup(vAddr uint64, PID vm.PID) (existloc int){
	existloc = -1
	for i = 0; i < len(lastvisit); i++{
		if vAddr == lastvisit[i].vaddr && PID == lastvisit[i].pid{
			existloc = i
			break
		}
	}
	//fmt.Println("=====existloc======", existloc )
	return existloc
}

// 在哪个gpu上
func (tlb *TLB) GetLastvisit(vAddr uint64, PID vm.PID) (existloc uint64){
	existloc = 0
	for i = 0; i < len(lastvisit); i++{
		if vAddr == lastvisit[i].vaddr && PID == lastvisit[i].pid{
			existloc = lastvisit[i].visitid
			break
		}
	}
	//fmt.Println("=====existloc======", existloc )
	return existloc
}

// 被某个gpu访问记录
func (tlb *TLB) Validinlocal(vAddr uint64, position uint64, PID vm.PID) bool{
	loc := tlb.Lastvisitlookup(vAddr, PID)
	if lastvisit[loc].remoteaccesstime[position] == 0{
		return false
	}
	return true
}

func (tlb *TLB) IsInLocal(vAddr uint64, position uint64, PID vm.PID) bool{
	return tlb.GetLastvisit(vAddr, PID) == position
}

func (tlb *TLB) UpdateLastvisit(vAddr uint64, position uint64, PID vm.PID) {	
	loc := tlb.Lastvisitlookup(vAddr, PID)
	lastvisit[loc].visitid = position

	for i = 1; i < numGPUs; i++{
		lastvisit[loc].remoteaccesstime[i] = 0
	}
	lastvisit[loc].remoteaccesstime[position] = 1
}

func (tlb *TLB) UpdateAccesstime(vAddr uint64, position uint64, loc int) {
	lastvisit[loc].remoteaccesstime[int(position)] ++
}

func (tlb *TLB) InsertLastVisit(vAddr uint64, position uint64, PID vm.PID) {

	for i = 1; i < numGPUs; i++{
		if i == int(position){
			remoteaccesstimes[i] = 1
		}else{
			remoteaccesstimes[i] = 0
		}	
	}
	tmp := Last{PID ,vAddr, position, remoteaccesstimes}
	lastvisit = append(lastvisit, tmp)	
}


func (tlb *TLB) CheckLastVisit(vAddr uint64, PID vm.PID) Last {
	loc := tlb.Lastvisitlookup(vAddr, PID)
	return lastvisit[loc]
}



func (tlb *TLB) Checkcounter(vAddr uint64, gid int, PID vm.PID) bool{
	loc := tlb.Lastvisitlookup(vAddr, PID)
	if lastvisit[loc].remoteaccesstime[gid] >= accesscounter{
		return true
	}
	return false
}



func (tlb *TLB) lookup(now sim.VTimeInSec) bool {
	msg := tlb.topPort.Peek()
	if msg == nil {
		//fmt.Println("======msgnil=======")
		return false
	}

	req := msg.(*vm.TranslationReq)


	
 
	wherevisit := tlb.Lastvisitlookup(req.VAddr, req.PID) 
	wherevisit = tlb.Lastvisitlookup(req.VAddr, req.PID)
	
	mshrEntry := tlb.mshr.Query(req.PID, req.VAddr) 
	if mshrEntry != nil {	
		return tlb.processTLBMSHRHit(now, mshrEntry, req)
	}

	setID := tlb.vAddrToSetID(req.VAddr)
	set := tlb.Sets[setID]
	wayID, page, found := set.Lookup(req.PID, req.VAddr)
	
	
	
	//if found && page.Valid{
		//------------------------------------------------------------
	/*
	if found && page.Valid && tlb.Validinlocal(req.VAddr, req.DeviceID) == true{
		if wherevisit != -1 &&  tlb.GetLastvisit(req.VAddr) != req.DeviceID{
			//found in TLB while need to remote access, no need for page table walk, only add pcie latency
			req.Remote = 1

			if tlb.Checkcounter(req.VAddr, int(req.DeviceID)) == true{
				//tlb.UpdateLastvisit(req.VAddr, req.DeviceID)
			 	req.Migrate = 1
			 }
			

			return tlb.handleTranslationMiss(now, req)
		}
		//tlb hit
		if strings.Contains(tlb.name, "L2"){
			localaccessGPU[req.DeviceID] ++
		}
		return tlb.handleTranslationHit(now, req, setID, wayID, page)
	}
	*/
	


	// 看看是不是在对应的l2 tlb里，如果没有就miss
	if strings.Contains(tlb.name, "GPU[1]"){
		gid = 1	
	}else if strings.Contains(tlb.name, "GPU[2]"){
		gid = 2
	}else if strings.Contains(tlb.name, "GPU[3]"){
		gid = 3
	}else if strings.Contains(tlb.name, "GPU[4]"){
		gid = 4
	}else if strings.Contains(tlb.name, "GPU[5]"){
		gid = 5
	}else if strings.Contains(tlb.name, "GPU[6]"){
		gid = 6
	}else if strings.Contains(tlb.name, "GPU[7]"){
		gid = 7
	}else if strings.Contains(tlb.name, "GPU[8]"){
		gid = 8
	}
	//fmt.Println(gid, "pid", req.PID)
	/*
	loc2 := tlb.Lastvisitlookup(req.VAddr)
	if loc2!=-1{
		fmt.Println("******req DeviceID",req.DeviceID, "lastvisit",lastvisit[loc2])
	}
	*/

	l2setID = int(req.VAddr / uint64(pagesize) % uint64(l2set))
	//fmt.Println(gid, req.PID, tlb.name)
	l2foundflag = tlb.LookupL2TLB(gid-1, req.VAddr, l2setID, req.PID) 

	/*
	loc9 := tlb.Lastvisitlookup(req.VAddr)
	if loc9 != -1 {
		fmt.Println("********l2", l2foundflag, req.VAddr, "pp",tlb.GetLastvisit(req.VAddr),"lv", lastvisit[loc9], "req", req.DeviceID)
	} else {

		fmt.Println("********l2", l2foundflag, req.VAddr, "pp",tlb.GetLastvisit(req.VAddr),"lv", -1, "req", req.DeviceID)
	}
	*/
	if strings.Contains(tlb.name, "L2") {
		if 	l2foundflag == -1{
			return tlb.handleTranslationMiss(now, req)
		} else {
			l2hit ++
			tlb.UpdateL2TLB(gid-1, req.VAddr, l2setID, l2foundflag) 
			return tlb.handleTranslationHit(now, req, setID, wayID, page)
		}
	}

	
	

	//fmt.Println("lookup page pos != req device", tlb.GetLastvisit(req.VAddr) != req.DeviceID) // almost false
	//fmt.Println("local",tlb.GetLastvisit(req.VAddr), req.DeviceID)
	if wherevisit != -1 &&  tlb.GetLastvisit(req.VAddr,req.PID) != req.DeviceID{
		
		//tlb.UpdateAccesstime(req.VAddr, req.DeviceID, wherevisit)
		//if hit in L1 miss in L2 still TLB hit
		if req.Remote == 1{
			req.Remote = 1
		}else{
			//miss in TLB and also need to remote access, need for page table walk, add pcie latency and ptw latency
			req.Remote = 2
		}

		/*
		//loc2 := tlb.Lastvisitlookup(req.VAddr)
		//fmt.Println("************\n+accesstime", lastvisit[loc2], "req", req.DeviceID, tlb.GetLastvisit(req.VAddr))
		tlb.UpdateAccesstime(req.VAddr, req.DeviceID, wherevisit)
		
		//loc3 := tlb.Lastvisitlookup(req.VAddr)
		//fmt.Println("+accesstime", lastvisit[loc3], "check", tlb.Checkcounter(req.VAddr, int(req.DeviceID)) )
			
		//if tlb.Checkcounter(req.VAddr, int(req.DeviceID)) == true{
		if tlb.Checkcounter(req.VAddr, int(req.DeviceID)) == true{
			//fmt.Println("page position", tlb.GetLastvisit(req.VAddr), "req vaddr", req.VAddr)
			//loc2 := tlb.Lastvisitlookup(req.VAddr)
			//fmt.Println("-req DeviceID",req.DeviceID, "lastvisit",lastvisit[loc2])

			req.Migrate = 1
			return tlb.handleTranslationMiss(now, req)

			//tlb.UpdateLastvisit(req.VAddr, req.DeviceID)
			//loc1 := tlb.Lastvisitlookup(req.VAddr)
			//fmt.Println( "+req DeviceID",req.DeviceID, "lastvisit",lastvisit[loc1])
			
		}
		*/

		
		//fmt.Println("======miss migrate=======", req.VAddr, req.Migrate, tlb.name)
	}

	if found && page.Valid {
		return tlb.handleTranslationHit(now, req, setID, wayID, page)
	}
	

	return tlb.handleTranslationMiss(now, req)
}

func (tlb *TLB) handleTranslationHit(
	now sim.VTimeInSec,
	req *vm.TranslationReq,
	setID, wayID int,
	page vm.Page,
) bool {
	ok := tlb.sendRspToTop(now, req, page)
	if !ok {
		return false
	}

	tlb.visit(setID, wayID)
	tlb.topPort.Retrieve(now)

	tracing.TraceReqReceive(req, tlb)
	tracing.AddTaskStep(tracing.MsgIDAtReceiver(req, tlb), tlb, "hit")
	tracing.TraceReqComplete(req, tlb)



	if strings.Contains(tlb.name, "L2TLB"){
		L2hit ++
	}


	
	
	// add/update access counter
	wherevisit := tlb.Lastvisitlookup(req.VAddr, req.PID) 
	if strings.Contains(tlb.name, "L1") {
		if wherevisit != -1 && tlb.GetLastvisit(req.VAddr, req.PID) != req.DeviceID {
			tlb.UpdateAccesstime(req.VAddr, req.DeviceID, wherevisit)
		}

		
		// 记得注释掉
		/*
		if wherevisit == -1{
			for i = 1; i < numGPUs; i++{
				if i == int(req.DeviceID){
					remoteaccesstimes[i] = 1
				}else{
					remoteaccesstimes[i] = 0
				}	
			}
			tmp := Last{req.VAddr, req.DeviceID, remoteaccesstimes}
			lastvisit = append(lastvisit, tmp)	
		}
		*/
		
	}
		

		
	return true
}

func (tlb *TLB) handleTranslationMiss(
	now sim.VTimeInSec,
	req *vm.TranslationReq,
) bool {

	if tlb.mshr.IsFull() {
		//fmt.Println("======false=======")
		return false
	}


	fetched := tlb.fetchBottom(now, req)
	//fmt.Println("======miss=======", req.Migrate)
	//fmt.Println("tlb miss req.Remote",req.Remote, tlb.name, req.VAddr)
	if fetched {
		tlb.topPort.Retrieve(now)
		tracing.TraceReqReceive(req, tlb)
		tracing.AddTaskStep(tracing.MsgIDAtReceiver(req, tlb), tlb, "miss")
		

		//fmt.Println("success",req.Remote, tlb.name, req.VAddr)
		if strings.Contains(tlb.name, "L2TLB") {
			if req.Remote == 1{
				L2hit ++
			}else{
				L2miss ++
			}
		}
		//fmt.Println("======req=======", tlb.name,req.VAddr)
		
		// add/update access counter
		wherevisit := tlb.Lastvisitlookup(req.VAddr, req.PID) 
		if strings.Contains(tlb.name, "L1") {
			if wherevisit != -1 && tlb.GetLastvisit(req.VAddr, req.PID) != req.DeviceID {
				tlb.UpdateAccesstime(req.VAddr, req.DeviceID, wherevisit)
			}
			/*
			if wherevisit == -1{
				for i = 1; i < numGPUs; i++{
					if i == int(req.DeviceID){
						remoteaccesstimes[i] = 1
					}else{
						remoteaccesstimes[i] = 0
					}	
				}
				tmp := Last{req.VAddr, req.DeviceID, remoteaccesstimes}
				lastvisit = append(lastvisit, tmp)	
			}
			*/
			
		}
		

		return true
	}

	//fmt.Println("======false=======")
	return false
}

func (tlb *TLB) vAddrToSetID(vAddr uint64) (setID int) {
	return int(vAddr / tlb.pageSize % uint64(tlb.numSets))
}

func (tlb *TLB) sendRspToTop(
	now sim.VTimeInSec,
	req *vm.TranslationReq,
	page vm.Page,
) bool {
	rsp := vm.TranslationRspBuilder{}.
		WithSendTime(now).
		WithSrc(tlb.topPort).
		WithDst(req.Src).
		WithRspTo(req.ID).
		WithPage(page).
		Build()

	err := tlb.topPort.Send(rsp)
	if err == nil {
		return true
	}

	return false
}

func (tlb *TLB) processTLBMSHRHit(
	now sim.VTimeInSec,
	mshrEntry *mshrEntry,
	req *vm.TranslationReq,
) bool {
	mshrEntry.Requests = append(mshrEntry.Requests, req)

	tlb.topPort.Retrieve(now)
	tracing.TraceReqReceive(req, tlb)
	tracing.AddTaskStep(tracing.MsgIDAtReceiver(req, tlb), tlb, "mshr-hit")

	if strings.Contains(tlb.name, "L2TLB"){
		if req.Remote != 1{
			L2mshr ++
		}
	}

	return true
}
 
func (tlb *TLB) fetchBottom(now sim.VTimeInSec, req *vm.TranslationReq) bool {
	fetchBottom := vm.TranslationReqBuilder{}.
		WithSendTime(now).
		WithSrc(tlb.bottomPort).
		WithDst(tlb.LowModule).
		WithPID(req.PID).
		WithVAddr(req.VAddr).
		WithDeviceID(req.DeviceID).
		WithRemote(req.Remote).
		WithMigrate(req.Migrate).
		WithGmmuhit(req.Gmmuhit).
		Build()
	err := tlb.bottomPort.Send(fetchBottom)
	if err != nil {
		//fmt.Println("======nil=======")
		return false
	}

	

		//fmt.Println("======req=======",req.Migrate)

		// count page sharing
	if strings.Contains(tlb.name, "GPU[1]"){
		gid = 1	
	}else if strings.Contains(tlb.name, "GPU[2]"){
		gid = 2
	}else if strings.Contains(tlb.name, "GPU[3]"){
		gid = 3
	}else if strings.Contains(tlb.name, "GPU[4]"){
		gid = 4
	}else if strings.Contains(tlb.name, "GPU[5]"){
		gid = 5
	}else if strings.Contains(tlb.name, "GPU[6]"){
		gid = 6
	}else if strings.Contains(tlb.name, "GPU[7]"){
		gid = 7
	}else if strings.Contains(tlb.name, "GPU[8]"){
		gid = 8
	}


	if strings.Contains(tlb.name, "L1"){
		pageInTable := tlb.Tablelookup(req.VAddr, req.PID)
		if pageInTable == -1{
			tlb.TableInsert(req.PID,gid, req.VAddr)
			table[len(table)-1].count ++

		}else{
			tlb.TableUpdate(gid, req.VAddr,req.PID)
			table[pageInTable].count ++
		}	
		for j = 0; j < 9; j++{
			sharePercentage[j] = 0
			accesstime[j] = 0
		} 
		for k = 0; k < len(table); k++{
			pagenumber = 0
			
			for j = 0; j < 9; j++{
				pagenumber += table[k].store[j]
			}
			sharePercentage[pagenumber-1] ++
			accesstime[pagenumber-1] += table[k].count			
		}
		
	}

	tlb.isL2 = strings.Contains(tlb.name, "L2")

	
	l2setID = int(req.VAddr / uint64(pagesize) % uint64(l2set))
	l2foundflag = tlb.LookupL2TLB(gid-1, req.VAddr, l2setID,req.PID) 


	if tlb.isL2{
		if l2foundflag != -1{
			//l2hit ++
			//tlb.UpdateL2TLB(gid-1, req.VAddr, l2setID, l2foundflag) 
		} else  {
			l2miss += 1 
			// l2 miss
			// l2tlb miss -> gmmu hit
			/*
			loc9 := tlb.Lastvisitlookup(req.VAddr)
			if loc9 != -1 {
				fmt.Println("********l2", l2foundflag, req.VAddr, "pp",tlb.GetLastvisit(req.VAddr),"lv", lastvisit[loc9], "req", req.DeviceID)
			} else {

				fmt.Println("********l2", l2foundflag, req.VAddr, "pp",tlb.GetLastvisit(req.VAddr),"lv", -1, "req", req.DeviceID)
			}
			*/
			if tlb.Lastvisitlookup(req.VAddr,req.PID) != -1 &&  tlb.GetLastvisit(req.VAddr,req.PID) == req.DeviceID{
				total_gmmu_hit += 1
				//req.Gmmuhit = 1
				//fmt.Println("gmmu hit !!!!!!",req.VAddr,tlb.GetLastvisit(req.VAddr),req.DeviceID, "gmmu hit" ,req.Gmmuhit)
				
			}
		}
		
	}

	


	/*
	fmt.Println("------------")
	for i = 0; i < len(l1l2table); i++{
		fmt.Println(l1l2table[i].store)
		//fmt.Println(table[i].vaddr )
	}
	*/

 	/*
	fmt.Println("------------")
	for i = 0; i < len(table); i++{
		//fmt.Println(table[i].store)
		//fmt.Println(table[i].vaddr )
		//fmt.Println(table[i].access)
	}
	*/
	
	
	fmt.Println("======page sharing=======", sharePercentage, accesstime)
	//fmt.Println("page sharing:", sharePercentage, " access time:", accesstime)
	

	
	mshrEntry := tlb.mshr.Add(req.PID, req.VAddr)
	mshrEntry.Requests = append(mshrEntry.Requests, req)
	mshrEntry.reqToBottom = fetchBottom

	tracing.TraceReqInitiate(fetchBottom, tlb,
		tracing.MsgIDAtReceiver(req, tlb))


	return true
}

// 把bottom里的req拿过来-> rsp, 一级一级往上传，startwalking里传回来的
func (tlb *TLB) parseBottom(now sim.VTimeInSec) bool {
	if tlb.respondingMSHREntry != nil {
		return false
	}

	item := tlb.bottomPort.Peek()
	if item == nil {
		//fmt.Println("======nilfalse=======")
		return false
	}


	rsp := item.(*vm.TranslationRsp)
	page := rsp.Page

	mshrEntryPresent := tlb.mshr.IsEntryPresent(rsp.Page.PID, rsp.Page.VAddr)
	if !mshrEntryPresent {
		tlb.bottomPort.Retrieve(now)
		return true
	}

	setID := tlb.vAddrToSetID(page.VAddr)
	set := tlb.Sets[setID]
	wayID, ok := tlb.Sets[setID].Evict()
	if !ok {
		panic("failed to evict")
	}
	set.Update(wayID, page)
	set.Visit(wayID)

	
	mshrEntry := tlb.mshr.GetEntry(rsp.Page.PID, rsp.Page.VAddr)
	tlb.respondingMSHREntry = mshrEntry
	mshrEntry.page = page

	tlb.mshr.Remove(rsp.Page.PID, rsp.Page.VAddr)
	tlb.bottomPort.Retrieve(now)
	tracing.TraceReqFinalize(mshrEntry.reqToBottom, tlb)

	
	// remove entry (l1 l2 tlb related to pid)-----------------------
	// lastvisit id 里device对应的page tlbentry
	// device id != lastvisit device id 

	/*
	if strings.Contains(tlb.name, "L1") || strings.Contains(tlb.name, "L2TLB"){
		
		set_id := tlb.vAddrToSetID(rsp.Page.VAddr)
		set2 := tlb.Sets[set_id]
		way_id, page2, found := set2.Lookup(rsp.Page.PID, rsp.Page.VAddr)
		way_id, page2, found = set2.Lookup(rsp.Page.PID, rsp.Page.VAddr)
		
		//fmt.Println(tlb.GetLastvisit(rsp.Page.VAddr), rsp.Page.DeviceID)

		if found && tlb.GetLastvisit(rsp.Page.VAddr) != rsp.Page.DeviceID {
			set2.DeletePage(way_id,page2)
			tlb.UpdateLastvisit(rsp.Page.VAddr, rsp.Page.DeviceID)
			
			//page2.Valid = false //确认一下page.valid代表什么
			//set.Update(way_id,page2)
		}
		

	}
	
	//-------------
	*/

	if strings.Contains(tlb.name, "GPU[1]"){
		gid = 1	
	}else if strings.Contains(tlb.name, "GPU[2]"){
		gid = 2
	}else if strings.Contains(tlb.name, "GPU[3]"){
		gid = 3
	}else if strings.Contains(tlb.name, "GPU[4]"){
		gid = 4
	}else if strings.Contains(tlb.name, "GPU[5]"){
		gid = 5
	}else if strings.Contains(tlb.name, "GPU[6]"){
		gid = 6
	}else if strings.Contains(tlb.name, "GPU[7]"){
		gid = 7
	}else if strings.Contains(tlb.name, "GPU[8]"){
		gid = 8
	}


	
	l2setID = int(rsp.Page.VAddr / uint64(pagesize) % uint64(l2set))
	l2foundflag = tlb.LookupL2TLB(gid-1, rsp.Page.VAddr, l2setID,rsp.Page.PID) 



	

	if strings.Contains(tlb.name, "L2"){
		if 	l2foundflag == -1{
			//l2miss += 1
			tlb.InsertL2TLB(gid-1, rsp.Page.VAddr, rsp.Page.PID,l2setID)
			
		}
	}

	


	//fmt.Println("l2miss",l2miss, "l2hit", l2hit, "gmmuhit", total_gmmu_hit)
	/*
	loc9 := tlb.Lastvisitlookup(rsp.Page.VAddr)
	if loc9 != -1 {
		fmt.Println("********l2", l2foundflag, rsp.Page.VAddr, "pp",tlb.GetLastvisit(rsp.Page.VAddr),"lv", lastvisit[loc9], "req", rsp.Page.DeviceID)
	} else {

		fmt.Println("********l2", l2foundflag, rsp.Page.VAddr, "pp",tlb.GetLastvisit(rsp.Page.VAddr),"lv", -1, "req", rsp.Page.DeviceID)
	}
	fmt.Println("Migrate or not", req.Migrate)
	*/
	

	

	return true
}

func (tlb *TLB) performCtrlReq(now sim.VTimeInSec) bool {
	item := tlb.controlPort.Peek()
	if item == nil {
		return false
	}

	item = tlb.controlPort.Retrieve(now)

	switch req := item.(type) {
	case *FlushReq:
		return tlb.handleTLBFlush(now, req)
	case *RestartReq:
		return tlb.handleTLBRestart(now, req)
	default:
		log.Panicf("cannot process request %s", reflect.TypeOf(req))
	}

	return true
}

func (tlb *TLB) visit(setID, wayID int) {
	set := tlb.Sets[setID]
	set.Visit(wayID)
}

func (tlb *TLB) handleTLBFlush(now sim.VTimeInSec, req *FlushReq) bool {
	rsp := FlushRspBuilder{}.
		WithSrc(tlb.controlPort).
		WithDst(req.Src).
		WithSendTime(now).
		Build()

	err := tlb.controlPort.Send(rsp)
	if err != nil {
		return false
	}

	for _, vAddr := range req.VAddr {
		setID := tlb.vAddrToSetID(vAddr)
		set := tlb.Sets[setID]
		wayID, page, found := set.Lookup(req.PID, vAddr)
		if !found {
			continue
		}

		page.Valid = false
		set.Update(wayID, page)
	}

	tlb.mshr.Reset()
	tlb.isPaused = true
	return true 
}

func (tlb *TLB) handleTLBRestart(now sim.VTimeInSec, req *RestartReq) bool {
	rsp := RestartRspBuilder{}.
		WithSendTime(now).
		WithSrc(tlb.controlPort).
		WithDst(req.Src).
		Build()

	err := tlb.controlPort.Send(rsp)
	if err != nil {
		return false
	}

	tlb.isPaused = false

	for tlb.topPort.Retrieve(now) != nil {
		tlb.topPort.Retrieve(now)
	}

	for tlb.bottomPort.Retrieve(now) != nil {
		tlb.bottomPort.Retrieve(now)
	}

	return true
}


func (tlb *TLB) Tablelookup(vAddr uint64,PID vm.PID) (existloc int){
	existloc = -1
	for i = 0; i < len(table); i++{
		if vAddr == table[i].vaddr && PID == table[i].pid {
			existloc = i
			break
		}
	}
	return existloc
}


func (tlb *TLB) TableInsert(PID vm.PID, gid int, vAddr uint64) bool{
	for i = 1; i < 9; i++{
		if i == gid{
			store[i] = 1
			access[i] = 1
		}else{
			store[i] = 0
			access[i] = 0
		}
	}
	tmp := Table{PID, vAddr, store, 0, access}
	table = append(table, tmp)
	return true
} 

func (tlb *TLB) TableUpdate(gid int, vAddr uint64,PID vm.PID) bool {
	for i = 0; i < len(table); i++{
		if table[i].vaddr == vAddr && table[i].pid == PID{
			table[i].store[gid] = 1
			table[i].access[gid] ++
		}
	}
	return true
}

func (tlb *TLB) TableDelete(vAddr uint64,PID vm.PID) bool{
	for i=0; i<len(table); i++{	
		if vAddr == table[i].vaddr && PID == table[i].pid{
			table = append(table[:i], table[i+1:]...)
			break
		}
	} 
	return true
}

/*
func (tlb *TLB) TableRemove(gid int, vAddr uint64) bool {
	for i = 0; i < len(table); i++{
		if table[i].vaddr == vAddr{
			table = append(table[:i], table[i+1:]...)
		}
	}
	return true
}
*/
func (tlb *TLB) FlushMMUTLB(VAddr uint64, PID vm.PID) bool{

	mmusetID = int(VAddr / setpagesize % uint64(mmuset))
	flush_pos := -1
	for i = 0; i < mmuway; i++{
		if mmutlb[mmusetID][i].addr == VAddr && mmutlb[mmusetID][i].pid == PID { 
			flush_pos = i
			break
		}
	}

	for j := flush_pos; j < mmuway-1; j++{
		mmutlb[mmusetID][j] = mmutlb[mmusetID][j+1]
	}

	tmp_iommu := *new(iommu)
	mmutlb[mmusetID][mmuway-1] = tmp_iommu


	return true
}

func (tlb *TLB) InsertMMUTLB(VAddr uint64, PID vm.PID) bool{
	mmusetID = int(VAddr / setpagesize % uint64(mmuset))	
	for i = 0; i < mmuway; i++{
		if mmutlb[mmusetID][i].addr == 0{
			mmutlb[mmusetID][i].addr = VAddr
			mmutlb[mmusetID][i].pid = PID
			//fmt.Println("======= insert into mmutlb =======", VAddr)
			break
		}
	}

	if i == mmuway{
		for k = 0; k < mmuway-1; k++ {
			mmutlb[mmusetID][k] = mmutlb[mmusetID][k+1]
		}
		mmutlb[mmusetID][mmuway-1].addr = VAddr	
		mmutlb[mmusetID][mmuway-1].pid = PID	
	}
	return true
}


func (tlb *TLB) UpdateMMUTLB(VAddr uint64, mmutlblocation int) bool{
	// move new access entry to list tail
	mmusetID = int(VAddr / setpagesize % uint64(mmuset))
	// compute mmutlb entry length
	for i = 0; i < mmuway; i++{
		if mmutlb[mmusetID][i].addr == 0{
			break
		}	
	}
	mmutlblength = i
	// pick out this entry and move to the list tail
	mmutlbtmp := iommu{mmutlb[mmusetID][mmutlblocation].pid,mmutlb[mmusetID][mmutlblocation].addr}
	for i = mmutlblocation; i < mmutlblength-1; i++{
		mmutlb[mmusetID][i] = mmutlb[mmusetID][i+1]
	}
	mmutlb[mmusetID][mmutlblength-1] = mmutlbtmp

	return true
}


func (tlb *TLB) LookupMMUTLB(VAddr uint64, mmusetID int, PID vm.PID)(existloc int) {
	existloc = -1
	for i = 0; i < mmuway; i++ {
		if mmutlb[mmusetID][i].addr == VAddr && mmutlb[mmusetID][i].pid == PID{
			existloc  = i
			break
		}
	}
	return existloc
}



func (tlb *TLB) InsertL2TLB(gid int, VAddr uint64, PID vm.PID, l2setID int) bool {
	
	//insert into L2
	for i = 0; i < l2way; i++ {
		if l2tlb[gid][l2setID][i].addr == 0{
			l2tlb[gid][l2setID][i].addr = VAddr
			l2tlb[gid][l2setID][i].pid = PID
			//fmt.Println("======= insert into l2 =======", gid, VAddr)
			break
		}
	}

	// if full, evict in L2
	if i == l2way {
		//fmt.Println("======= l2 full=======", gid, VAddr)
		//evictl2tlb := l2{l2tlb[gid][l2setID][0].pid,l2tlb[gid][l2setID][0].addr}
		for j = 0; j < l2way-1; j++ {
				l2tlb[gid][l2setID][j] = l2tlb[gid][l2setID][j+1]
		}
		l2tlb[gid][l2setID][l2way-1].addr=VAddr	
		l2tlb[gid][l2setID][l2way-1].pid=PID	

		//if evict in local gpu, update in table
		//tlb.TableDelet(gid, evictl2tlb)
		//fmt.Println("======= remove=======", evictl2tlb)
		// if not exist in anywhere, insert into iommu tlb	s
		
			
	}
	return true
}

func (tlb *TLB) UpdateL2TLB(gid int, VAddr uint64, l2setID int, l2tlblocation int) bool{
	// move new access entry to list tail
	// compute l2tlb entry length
	for i = 0; i < l2way; i++{
		if l2tlb[gid][l2setID][i].addr == 0{
			break
		}
	}
	l2tlblength = i
	// pick out this entry and move to the list tail
	l2tlbtmp := l2{l2tlb[gid][l2setID][l2tlblocation].pid,l2tlb[gid][l2setID][l2tlblocation].addr}
	for i = l2tlblocation; i < l2tlblength-1; i++{
		l2tlb[gid][l2setID][i] = l2tlb[gid][l2setID][i+1]
	}
	l2tlb[gid][l2setID][l2tlblength-1] = l2tlbtmp

	//fmt.Println("update l2" )
	return true
}


func (tlb *TLB) LookupL2TLB(gid int, VAddr uint64, l2setID int, PID vm.PID)(existloc int) {
	existloc = -1
	for i = 0; i < l2way; i++ {
		if l2tlb[gid][l2setID][i].addr == VAddr && l2tlb[gid][l2setID][i].pid == PID{
			existloc  = i
			break
		}
	}
	return existloc
}


func (tlb *TLB) DeleteL2TLB(gid int, VAddr uint64, PID vm.PID) bool {
	l2setID = int(VAddr/ uint64(pagesize) % uint64(l2set))
	//fmt.Println("@@@@@@@@@@@@@@@@@@@@@@@@@@@@@@", gid, l2setID, l2way)
	for i = 0; i < l2way; i++ {
		if l2tlb[gid][l2setID][i].addr == VAddr && l2tlb[gid][l2setID][i].pid == PID{
			break
		}
	}

	delete_index := i 
	for i = delete_index; i < l2way-1; i++{
		l2tlb[gid][l2setID][i] = l2tlb[gid][l2setID][i+1]
	}

	return true
	
}

