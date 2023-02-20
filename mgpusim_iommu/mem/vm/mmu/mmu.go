package mmu
// iommu----------------


import (
	"log"
	"reflect" 
	"fmt"
	"strconv"
	"sort"
	"gitlab.com/akita/akita/v2/sim"
	"gitlab.com/akita/mem/v2/vm"
	"gitlab.com/akita/util/v2/akitaext"
	"gitlab.com/akita/util/v2/tracing"
	"gitlab.com/akita/mem/v2/vm/addresstranslator"
	"gitlab.com/akita/util/v2/ca"
	"gitlab.com/akita/mem/v2/vm/tlb"
)

type iommu struct{
	pid    ca.PID
	addr   uint64
}

type addrNpid struct{
	pid    ca.PID
	addr   uint64
}


func convertToBin(num int) string {
	s := ""
	for ; num > 0; num /= 2{
		lsb := num % 2
		s = strconv.Itoa(lsb) + s
	}
	return s
}
 

var 
(	
	UVM_BATCH_FIXED_LATENCY int=1000
	ENABLE_DIRECT_OS		bool=true// false = iommu, true = ideal
	setpagesize				uint64 = 4096
	mmuset              	int = 64
	mmuway             	 	int          
	mmutlb              	[64][32]iommu
	numGPU                  int = 5
	mmusetID           		int
	mmutlblength        	int
	mmutlblocation      	int
	foundflag           	int
	foundflagtmp			int
	tablefoundflag      	int
	existloc               	int
	i,j,k,p                 int
	iommutlb             	*tlb.TLB 
	page_walk_cache         [9][128] string
	slides[5]              	string
	page_loc               	int
	cache_level            	int
	fullFlag           		int
	cache_entry 			int = 128
	hit_loc 				int
	hit_other_loc           int
	IOMMU_PWC[128]          string
	iommu_cache_entry       int = 128
	otherPWClevel           int
	iommu_cache_level       int
	hitotherperc[25]        int 
	allPWChit               int
	allotherperc[25]        int 
	evictloc                int  
	max                     int
	cyclecount              int = 0
	hitrate[5]              int
	remotehitrate[5]        int
	io_hitrate[6]           int  
	localqueue				[9][]uint64
	iommuqueue              []uint64 
	iommuPTW     			int = 16
	localPTW 				int = 8
	t,l                     int 
	iommuptw_wait        	int = 0
	iommuptw_wait_total		int = 0

	localptw_wait[9]        int 
	localptw_wait_total[9]  int

	remoteptw_wait          int = 0
	remote_benefit          int
	local_iommu_benefit     int
	tmpiommuptw             []int
	tmplocalptw             [9][]int
	tmpiommu                []int
	tmplocal                [9][]int
	locallatency, iommulatency, remotelatency  int
	actuallatency			int=0
	ioactuallatency			int=0
	

	newtmpiommu                []int
	newtmplocal                [9][]int
	requestPTW[9]			int 
	invalidPTW[9]			int 
	unnecessary[9]			int 



	localwaittotal          int=0
	iommuwaittotal          int=0
	pagefalttotal           int
	local_memory_access     int=0
	remotewaittotal			int
	localwaittimes          int
	iommuwaittimes          int
	pwctimes                int
	pagefaulttimes          int
	remotewaittimes			int
	localwaitportion        int
	iommuwaitportion        int
	pagefaultportion        int
	pwcportion              int
	remotewaitportion		int
	iopwcportion 			int
	io_memory_access_latency				int=0
	iopwctimes				int
	lastaccess				bool=false
	lastaccesstmp			bool
	pcieall			int=0
	replayall		int=0

	migrationtimes  		int = 0
	pagenotequal 			int = 0
	pin						int = 0

	translatecount			*addresstranslator.AddressTranslator

	pcie 					int = 150
	needmigrateflg   				int
	validationflg			int

	totallatency   			int64 = 0 
	invalidlatency      	int64 = 0
	iototallatency			int64 = 0
	pflocallatency			int64 = 0
	waitptw					int64 = 0
	resolvelatency			int64 = 0
	localwaitlatency	    int	= 0

	cache_level_invalid		int
	invalidRequestlatency			int
	remoteaccessLatency    	int
	remoteaccessGPU[9]      int
	localaccessGPU[9]		int

	each_pwc_level_latency  int=200

	// host side
	hostPageTable 			vm.PageTable=vm.NewPageTable(12)
	tmp_os_flag				bool
	hostThread				int=32

	// iommu fake tlb, used to decide batch
	iommuPageTable 			vm.PageTable=vm.NewPageTable(12)

	// batch related 
	BATCH_SIZE_MAX			int=256
	faultBatch				[]transaction
	batchWaitingCnt			int=0
	batchThreshold			int=512
	fault_num				int=0
	batch_cycle				int=0
	batch_num				int=0
	batch_num_helper		int=0

	// to check
	req_gointo_batch		int=0
	ave_waiting_time_of_batch int=0
	total_waiting_time_of_batch int=0
	each_batch_waiting			int=0
	iommu_only_latency			int=0

	ultimate_total_latency		int=0
	ultimate_total_requests		int=0
	req_gointo_iommu			int=0
	req_hitin_gmmu				int=0
	req_hitin_iommu 			int=0
	req_hitin_iommutlb			int=0
	req_hitin_iommupwc			int=0
	req_hitin_iommupt 			int=0
	total_compulsory_miss 		int=0
	total_inothergpu_miss		int=0
	total_local_wait_queue_len	int=0
	total_io_wait_queue_len		int=0
	startWalking_round			int=0
	ultimate_os_wait           int=0

	

	iommu_pwc_latency			int=0
	req_hitin_iommupwc_and_pt   int=0
	saved_latency_by_directos	int=0


	test_tmp				int=0
	test_tmp2 				int=0

	// gmmu pagetable
	GmmuPageTable			[8]vm.PageTable=[8]vm.PageTable{vm.NewPageTable(12),vm.NewPageTable(12),vm.NewPageTable(12),vm.NewPageTable(12), vm.NewPageTable(12),vm.NewPageTable(12),vm.NewPageTable(12),vm.NewPageTable(12)}
	
	// cold or hot miss record table, hot miss: request time > 1
	request_record 			  map[addrNpid]int=make(map[addrNpid] int)
	iommu_cachelevel_latency  map[int]int=make(map[int] int)

	// check, multi application
	eachgpu_total_requests		[9]int 
	eachgpu_hit_gmmu			[9]int			
	eachgpu_goto_iommu			[9]int
	eachgpu_goto_os			[9]int
	eachgpu_hit_io_tlb			[9]int
	eachgpu_hit_io_pwc			[9]int
	eachgpu_hit_io_pt			[9]int
	eachgpu_hit_io_total			[9]int

	eachgpu_local_wait_latency			[9]int
	eachgpu_local_pwc_latency			[9]int
	eachgpu_pcieall_latency			[9]int
	eachgpu_io_wait_latency				[9]int
	eachgpu_io_pwc_latency 				[9]int
	eachgpu_oshandle_latency			[9]int
	eachgpu_alllatency			[9]int
	eachgpu_replayall_latency			[9]int
	eachgpu_compulsory_miss				[9]int
	eachgpu_inothergpu_miss				[9]int
	eachgpu_os_wait_latency				[9]int


)



type transaction struct {
	req       *vm.TranslationReq
	page      vm.Page
	cycleLeft int
	migration *vm.PageMigrationReqToDriver
	local     int
	locallevel     	int
	iolevel        	int
	localwait      	int
	iowait         	int
	oswait          int
	validflg		int      //validflag=0 -> normal ptw request, =1 -> invalidation request
	validid   		int
	remote			int
	os_handle_flag	bool
	direct_os_flag bool
	iommu_tlb_hit_flag  int
	iommupt_hit_flag  int

}

// MMU is the default mmu implementation. It is also an akita Component.
type MMU struct {
	sim.TickingComponent

	topPort       sim.Port
	migrationPort sim.Port

	MigrationServiceProvider sim.Port

	topSender akitaext.BufferedSender

	pageTable           vm.PageTable
	latency             int
	maxRequestsInFlight int

	walkingTranslations      []transaction
	migrationQueue           []transaction
	migrationQueueSize       int
	currentOnDemandMigration transaction
	isDoingMigration         bool

	toRemoveFromPTW        []int
	PageAccessedByDeviceID map[uint64][]uint64
}



// Tick defines how the MMU update state each cycle
func (mmu *MMU) Tick(now sim.VTimeInSec) bool {
	madeProgress := false

	madeProgress = mmu.topSender.Tick(now) || madeProgress
	madeProgress = mmu.sendMigrationToDriver(now) || madeProgress
	madeProgress = mmu.walkPageTable(now) || madeProgress
	madeProgress = mmu.processMigrationReturn(now) || madeProgress
	madeProgress = mmu.parseFromTop(now) || madeProgress

	return madeProgress
}

func (mmu *MMU) trace(now sim.VTimeInSec, what string) {
	ctx := sim.HookCtx{
		Domain: mmu,
		Now:    now,
		Item:   what,
	}

	mmu.InvokeHook(ctx)
}

func (mmu *MMU) walkPageTable(now sim.VTimeInSec) bool {
	madeProgress := false

	// reset batchwaitingcnt and calculate total wait
	no_batch_flag := true
	for i := 0; i < len(mmu.walkingTranslations); i++{
		if mmu.walkingTranslations[i].os_handle_flag == true{
			no_batch_flag = false
		}
	}
	if no_batch_flag == true {
		total_waiting_time_of_batch += batchWaitingCnt
		batchWaitingCnt = 0
		fault_num = 0
	}
	//fmt.Println("************************************************out loop", len(mmu.walkingTranslations))
	for i := 0; i < len(mmu.walkingTranslations); i++ {
		//fmt.Println("######################inner loop", len(mmu.walkingTranslations))

		if mmu.walkingTranslations[i].os_handle_flag == true && fault_num < BATCH_SIZE_MAX && batchWaitingCnt < batchThreshold {
			batchWaitingCnt += 1
			madeProgress = true
			batch_num_helper = 2
			//fmt.Println("--------i", i)
			//fmt.Println(len(mmu.walkingTranslations))

		}else {
			//fmt.Println(fault_num, batchWaitingCnt, batch_num_helper, no_batch_flag)
			if  batch_num_helper == 2 && (fault_num >= BATCH_SIZE_MAX || batchWaitingCnt >= batchThreshold) {
				batch_num += 1
				batch_num_helper = 1
			}

			
			if mmu.walkingTranslations[i].cycleLeft > 0 {
				mmu.walkingTranslations[i].cycleLeft--
				madeProgress = true
				//fmt.Println("walking translationInPipeline", mmu.walkingTranslations[i])
				continue
			}
			if mmu.walkingTranslations[i].validflg == 0 {
				madeProgress = mmu.finalizePageWalk(now, i) || madeProgress
			}
			if mmu.walkingTranslations[i].validflg == 1{
				mmu.walkingTranslations = append(mmu.walkingTranslations[:i], mmu.walkingTranslations[i+1:]...)
			}
		}
	}

	tmp := mmu.walkingTranslations[:0]
	for i := 0; i < len(mmu.walkingTranslations); i++ {
		if !mmu.toRemove(i) {
			tmp = append(tmp, mmu.walkingTranslations[i])
		}
	}
	mmu.walkingTranslations = tmp
	mmu.toRemoveFromPTW = nil

	return madeProgress

}

func (mmu *MMU) finalizePageWalk(
	now sim.VTimeInSec,
	walkingIndex int,
) bool {
	req := mmu.walkingTranslations[walkingIndex].req
	page, found := mmu.pageTable.Find(req.PID, req.VAddr)
	lastaccess =false
	
	if !found {
		panic("page not found")
	}

	Bin := convertToBin(int(req.VAddr))
	// for len(Bin) < 57{
	// 	Bin = "0" + Bin[0:]
	// }
	
    // slides[0] = Bin[0:9]
	// slides[1] = Bin[0:18]
	// slides[2] = Bin[0:27]
	// slides[3] = Bin[0:36]

	for len(Bin) < 35{
  		Bin = "0" + Bin[0:]
 	}
 	//fmt.Println(Bin)
 	//slide_len := 3
 	slides[0] = Bin[0:9]
 	slides[1] = Bin[0:11]
 	slides[2] = Bin[0:19]
 	slides[3] = Bin[0:23]

	// Get local cache level
	
	cache_level = -2
	hit_loc = -1
	hit_other_loc = -1
	otherPWClevel = -2

	//check if valid in GMMU
	if iommutlb.Lastvisitlookup(req.VAddr, req.PID) != -1{
		lastaccess = iommutlb.Validinlocal(req.VAddr, req.DeviceID, req.PID)
	}

	for k = 3; k >= 0; k--{
		for i = 0; i < cache_entry; i++{
			if slides[k] == page_walk_cache[req.DeviceID][i]{
				cache_level = k
				hit_loc = i
				break
			}
		}
		if cache_level != -2{
			break 
		}
	}
	
	
	// update page walk cache
	// ir request.remote = 1 indicates the request is only need remote acess, no need ptw
	if req.Remote != 1 && cache_level != -2{ 
	//if iommutlb.Lastvisitlookup(req.VAddr) != -1 && lastaccess == true && cache_level != -2{
		if hit_loc != -1{
			for i = 0; i < cache_entry; i++{
				if page_walk_cache[req.DeviceID][i] == ""{
					break
				}
			}
			cache_length := i
			cache_tmp := page_walk_cache[req.DeviceID][hit_loc]
			for j = hit_loc; j < cache_length-1; j++{
				page_walk_cache[req.DeviceID][j] = page_walk_cache[req.DeviceID][j+1]
			}
			page_walk_cache[req.DeviceID][cache_length-1] = cache_tmp
		}
	}

	
	if req.Remote != 1 {
		//if iommutlb.Lastvisitlookup(req.VAddr) != -1 && lastaccess == true {
		if iommutlb.Lastvisitlookup(req.VAddr, req.PID) != -1 && req.Gmmuhit == 1{
			mmu.insertPWC(req.DeviceID, req.VAddr, cache_level)
		}else{
			mmu.insertPWC(req.DeviceID, req.VAddr, cache_level)
			// ---------
			//wherevisit := iommutlb.Lastvisitlookup(req.VAddr) 
			//iommutlb.UpdateAccesstime(req.VAddr, req.DeviceID, wherevisit)

			// insert IOMMU
			foundflag = -1
			mmusetID = int(req.VAddr / setpagesize % uint64(mmuset))
			foundflag = iommutlb.LookupMMUTLB(req.VAddr, mmusetID, req.PID)
			
			// iommutlb记录里有，需要更新或者flush
			if foundflag != -1{
				if req.Migrate == 1 {
					iommutlb.FlushMMUTLB(req.VAddr, req.PID)
				} else {
					iommutlb_index := iommutlb.LookupMMUTLB(req.VAddr, mmusetID, req.PID) 
					iommutlb.UpdateMMUTLB(req.VAddr, iommutlb_index)
				}
			}

			
			if foundflag == -1 && req.Gmmuhit != 1 && mmu.walkingTranslations[walkingIndex].direct_os_flag == false{
				iommutlb.InsertMMUTLB(req.VAddr, req.PID)
				//fmt.Println("insert iommu cache!!!")
				// walk IOMMU page table
				//check if in page walk cache, 
				//0-28 level 3, 0-21 level 2, 0-14 level 1, 0-7 level 0, not exist -2
				cache_level = -2
				hit_loc = -1
				for k = 3; k >= 0; k--{
					for i = 0; i < iommu_cache_entry; i++{
						if slides[k] == IOMMU_PWC[i]{
							cache_level = k
							hit_loc = i
							break
						}
					}
					if cache_level != -2{
						break 
					}
				}
			
				// update cache hit loc
				if hit_loc != -1{
					for i = 0; i < iommu_cache_entry; i++{
						if IOMMU_PWC[i] == ""{
							break
						}
					}
					cache_length := i
					cache_tmp := IOMMU_PWC[hit_loc]
					for j = hit_loc; j < cache_length-1; j++{
						IOMMU_PWC[j] = IOMMU_PWC[j+1]
					}
					IOMMU_PWC[cache_length-1] = cache_tmp
				}  
				
				// Insert into cache
				// if == -2, insert all 4 slides
				if cache_level == -2{
					//mmu.latency = 1960 + 955
					fullFlag = -1
					for i = 0; i < iommu_cache_entry; i++{
						if IOMMU_PWC[i] == ""{
							fullFlag = 99
							if i < iommu_cache_entry - 3{
								IOMMU_PWC[i] = slides[0]
								IOMMU_PWC[i+1] = slides[1]
								IOMMU_PWC[i+2] = slides[2]
								IOMMU_PWC[i+3] = slides[3]
								break
							}else if i == iommu_cache_entry - 3{
								IOMMU_PWC[i] = slides[0]
								IOMMU_PWC[i+1] = slides[1]
								IOMMU_PWC[i+2] = slides[2]
								IOMMU_PWC[0] = slides[3]
								break
							}else if i == iommu_cache_entry - 2{
								IOMMU_PWC[i] = slides[0]
								IOMMU_PWC[i+1] = slides[1]
								IOMMU_PWC[0] = slides[2]
								IOMMU_PWC[1] = slides[3]
								break
							}else if i == iommu_cache_entry - 1{
								IOMMU_PWC[i] = slides[0]
								IOMMU_PWC[0] = slides[1]
								IOMMU_PWC[1] = slides[2]
								IOMMU_PWC[2] = slides[3]
								break
							}		
						}
					}
					// if cache is full
					if fullFlag == -1{
						for j = 0; j < iommu_cache_entry-4; j++{
							IOMMU_PWC[j] = IOMMU_PWC[j+4]	
						}
						IOMMU_PWC[iommu_cache_entry-4] = slides[0]
						IOMMU_PWC[iommu_cache_entry-3] = slides[1]
						IOMMU_PWC[iommu_cache_entry-2] = slides[2]
						IOMMU_PWC[iommu_cache_entry-1] = slides[3]
					}			
				}	
					
				// find in l4 cache, insert 3 slides
				if cache_level == 0{
					//mmu.latency = 1773 + 955
					fullFlag = -1
					for i = 0; i < iommu_cache_entry; i++{
						if IOMMU_PWC[i] == ""{
							fullFlag = 99
							if i < iommu_cache_entry-2 {
								IOMMU_PWC[i] = slides[1]
								IOMMU_PWC[i+1] = slides[2]
								IOMMU_PWC[i+2] = slides[3]
								break
							}else if i == iommu_cache_entry-2{
								IOMMU_PWC[i] = slides[1]
								IOMMU_PWC[i+1] = slides[2]
								IOMMU_PWC[0] = slides[3]
								break
							}else if i == iommu_cache_entry-1{
								IOMMU_PWC[i] = slides[1]
								IOMMU_PWC[0] = slides[2]
								IOMMU_PWC[1] = slides[3]
								break
							}
						}
					}
					// if cache is full
					if fullFlag == -1{
						for j = 0; j < iommu_cache_entry-3; j++{
							IOMMU_PWC[j] = IOMMU_PWC[j+3]	
						}
						IOMMU_PWC[iommu_cache_entry-3] = slides[1]
						IOMMU_PWC[iommu_cache_entry-2] = slides[2]
						IOMMU_PWC[iommu_cache_entry-1] = slides[3]
					}			
				}		
					
				// find in L3 cache, insert 2 slides
				if cache_level == 1{
					//mmu.latency = 1586 + 955
					fullFlag = -1
					for i = 0; i < iommu_cache_entry; i++{
						if IOMMU_PWC[i] == ""{
							fullFlag = 99
							if i < iommu_cache_entry-1{
								IOMMU_PWC[i] = slides[2]
								IOMMU_PWC[i+1] = slides[3]
								break
							}else if i == iommu_cache_entry-1{
								IOMMU_PWC[i] = slides[2]
								IOMMU_PWC[0] = slides[3]
								break
							}
						}
					}
					if fullFlag == -1{
						for j = 0; j < iommu_cache_entry-2; j++{
							IOMMU_PWC[j] = IOMMU_PWC[j+2]		
						}
						IOMMU_PWC[iommu_cache_entry-2] = slides[2]
						IOMMU_PWC[iommu_cache_entry-1] = slides[3]
					}			
				}
			
				// found in l2 cache, insert 1 slide1
				if cache_level == 2{
					//mmu.latency = 1399 + 955
					fullFlag = -1
					for i = 0; i < iommu_cache_entry; i++{
						if IOMMU_PWC[i] == ""{
							fullFlag = 99
							IOMMU_PWC[i] = slides[3]
							break
						}
					}
					if fullFlag == -1{
						for j = 0; j < iommu_cache_entry-1; j++{
							IOMMU_PWC[j] = IOMMU_PWC[j+1]	
						}
						IOMMU_PWC[iommu_cache_entry-1] = slides[3]
					}				
				}
			}
		}
	}
		

	mmu.walkingTranslations[walkingIndex].page = page

	if page.IsMigrating {
		return mmu.addTransactionToMigrationQueue(walkingIndex)
	}

	if mmu.pageNeedMigrate(mmu.walkingTranslations[walkingIndex]) {
		return mmu.addTransactionToMigrationQueue(walkingIndex)
	}

	

	return mmu.doPageWalkHit(now, walkingIndex)
}

func (mmu *MMU) addTransactionToMigrationQueue(walkingIndex int) bool {
	if len(mmu.migrationQueue) >= mmu.migrationQueueSize {
		return false
	}

	mmu.toRemoveFromPTW = append(mmu.toRemoveFromPTW, walkingIndex)
	mmu.migrationQueue = append(mmu.migrationQueue,
		mmu.walkingTranslations[walkingIndex])

	page := mmu.walkingTranslations[walkingIndex].page
	page.IsMigrating = true
	mmu.pageTable.Update(page)

	return true
}

func (mmu *MMU) insertPWC (deviceID uint64, VAddr uint64, cache_level int){
	//implement decimal to binary
	Bin := convertToBin(int(VAddr))
	//fmt.Println(Bin)
	// for len(Bin) < 57{
	// 	Bin = "0" + Bin[0:]
	// }
	
    // slides[0] = Bin[0:9]
	// slides[1] = Bin[0:18]
	// slides[2] = Bin[0:27]
	// slides[3] = Bin[0:36]

	for len(Bin) < 35{
  		Bin = "0" + Bin[0:]
 	}
 	//fmt.Println(Bin)
 	//slide_len := 3
 	slides[0] = Bin[0:9]
 	slides[1] = Bin[0:11]
 	slides[2] = Bin[0:19]
 	slides[3] = Bin[0:23]


   /* for i = 2; i < 4 ; i++{
			slides[i] = Bin[0:slide_len*(i-1)+12]
	}*/

	// Insert into cache
	// if == -2, insert all 4 slides
	if cache_level == -2{
		fullFlag = -1
		for i = 0; i < cache_entry; i++{
			if page_walk_cache[deviceID][i] == ""{
				fullFlag = 99
				if i < cache_entry - 3{
					page_walk_cache[deviceID][i] = slides[0]
					page_walk_cache[deviceID][i+1] = slides[1]
					page_walk_cache[deviceID][i+2] = slides[2]
					page_walk_cache[deviceID][i+3] = slides[3]
					break
				}else if i == cache_entry - 3{
					page_walk_cache[deviceID][i] = slides[0]
					page_walk_cache[deviceID][i+1] = slides[1]
					page_walk_cache[deviceID][i+2] = slides[2]
					page_walk_cache[deviceID][0] = slides[3]
					break
				}else if i == cache_entry - 2{
					page_walk_cache[deviceID][i] = slides[0]
					page_walk_cache[deviceID][i+1] = slides[1]
					page_walk_cache[deviceID][0] = slides[2]
					page_walk_cache[deviceID][1] = slides[3]
					break
				}else if i == cache_entry - 1{
					page_walk_cache[deviceID][i] = slides[0]
					page_walk_cache[deviceID][0] = slides[1]
					page_walk_cache[deviceID][1] = slides[2]
					page_walk_cache[deviceID][2] = slides[3]
					break
				}		
			}
		}
		// if cache is full
		if fullFlag == -1{
			for j = 0; j < cache_entry-4; j++{
				page_walk_cache[deviceID][j] = page_walk_cache[deviceID][j+4]	
			}
			page_walk_cache[deviceID][cache_entry-4] = slides[0]
			page_walk_cache[deviceID][cache_entry-3] = slides[1]
			page_walk_cache[deviceID][cache_entry-2] = slides[2]
			page_walk_cache[deviceID][cache_entry-1] = slides[3]
		}			
	}

	// find in l5 cache, insert 3 slides
	if cache_level == 0{
		fullFlag = -1
		for i = 0; i < cache_entry; i++{
			if page_walk_cache[deviceID][i] == ""{
				fullFlag = 99
				if i < cache_entry-2 {
					page_walk_cache[deviceID][i] = slides[1]
					page_walk_cache[deviceID][i+1] = slides[2]
					page_walk_cache[deviceID][i+2] = slides[3]
					break
				}else if i == cache_entry-2{
					page_walk_cache[deviceID][i] = slides[1]
					page_walk_cache[deviceID][i+1] = slides[2]
					page_walk_cache[deviceID][0] = slides[3]
					break
				}else if i == cache_entry-1{
					page_walk_cache[deviceID][i] = slides[1]
					page_walk_cache[deviceID][0] = slides[2]
					page_walk_cache[deviceID][1] = slides[3]
					break
				}
			}
		}
		// if cache is full
		if fullFlag == -1{
			for j = 0; j < cache_entry-3; j++{
				page_walk_cache[deviceID][j] = page_walk_cache[deviceID][j+3]	
			}
			page_walk_cache[deviceID][cache_entry-3] = slides[1]
			page_walk_cache[deviceID][cache_entry-2] = slides[2]
			page_walk_cache[deviceID][cache_entry-1] = slides[3]
		}			
	}

	// find in L4 cache, insert 2 slides
	if cache_level == 1{
		fullFlag = -1
		for i = 0; i < cache_entry; i++{
			if page_walk_cache[deviceID][i] == ""{
				fullFlag = 99
				if i < cache_entry-1{
					page_walk_cache[deviceID][i] = slides[2]
					page_walk_cache[deviceID][i+1] = slides[3]
					break
				}else if i == cache_entry-1{
					page_walk_cache[deviceID][i] = slides[2]
					page_walk_cache[deviceID][0] = slides[3]
					break
				}
			}
		}
		if fullFlag == -1{
			for j = 0; j < cache_entry-2; j++{
				page_walk_cache[deviceID][j] = page_walk_cache[deviceID][j+2]	
			}
			page_walk_cache[deviceID][cache_entry-2] = slides[2]
			page_walk_cache[deviceID][cache_entry-1] = slides[3]
		}			
	}

	// found in l3 cache, insert 1 slide1
	if cache_level == 2{
		fullFlag = -1
		for i = 0; i < cache_entry; i++{
			if page_walk_cache[deviceID][i] == ""{
				fullFlag = 99
				page_walk_cache[deviceID][i] = slides[3]
				break
			}
		}
		if fullFlag == -1{
			for j = 0; j < cache_entry-1; j++{
				page_walk_cache[deviceID][j] = page_walk_cache[deviceID][j+1]	
			}
			page_walk_cache[deviceID][cache_entry-1] = slides[3]
		}				
	}	
	// found in l2 cache, one memory access	
}

func (mmu *MMU) pageNeedMigrate(walking transaction) bool {
	
	if walking.req.DeviceID == walking.page.DeviceID {
		return false
	}

	if !walking.page.Unified {
		return false
	}

	if walking.page.IsPinned {
		return false
	}

	return true
}

func (mmu *MMU) doPageWalkHit(
	now sim.VTimeInSec,
	walkingIndex int,
) bool {
	if !mmu.topSender.CanSend(1) {
		return false
	}
	walking := mmu.walkingTranslations[walkingIndex]

	rsp := vm.TranslationRspBuilder{}.
		WithSendTime(now).
		WithSrc(mmu.topPort).
		WithDst(walking.req.Src).
		WithRspTo(walking.req.ID).
		WithPage(walking.page).
		Build()

	mmu.topSender.Send(rsp)
	mmu.toRemoveFromPTW = append(mmu.toRemoveFromPTW, walkingIndex)

	tracing.TraceReqComplete(walking.req, now, mmu)

	return true
}

func (mmu *MMU) sendMigrationToDriver(
	now sim.VTimeInSec,
) (madeProgress bool) {
	if len(mmu.migrationQueue) == 0 {
		return false
	}

	trans := mmu.migrationQueue[0]
	req := trans.req
	page, found := mmu.pageTable.Find(req.PID, req.VAddr)
	if !found {
		panic("page not found")
	}
	trans.page = page

	if req.DeviceID == page.DeviceID || page.IsPinned {
		mmu.sendTranlationRsp(now, trans)
		mmu.migrationQueue = mmu.migrationQueue[1:]
		mmu.markPageAsNotMigratingIfNotInTheMigrationQueue(page)

		return true
	}

	if mmu.isDoingMigration {
		return false
	}

	migrationInfo := new(vm.PageMigrationInfo)
	migrationInfo.GPUReqToVAddrMap = make(map[uint64][]uint64)
	migrationInfo.GPUReqToVAddrMap[trans.req.DeviceID] =
		append(migrationInfo.GPUReqToVAddrMap[trans.req.DeviceID],
			trans.req.VAddr)

	mmu.PageAccessedByDeviceID[page.VAddr] =
		append(mmu.PageAccessedByDeviceID[page.VAddr], page.DeviceID)

	migrationReq := vm.NewPageMigrationReqToDriver(
		now, mmu.migrationPort, mmu.MigrationServiceProvider)
	migrationReq.PID = page.PID
	migrationReq.PageSize = page.PageSize
	migrationReq.CurrPageHostGPU = page.DeviceID
	migrationReq.MigrationInfo = migrationInfo
	migrationReq.CurrAccessingGPUs = unique(mmu.PageAccessedByDeviceID[page.VAddr])
	migrationReq.RespondToTop = true

	err := mmu.migrationPort.Send(migrationReq)
	if err != nil {
		return false
	}

	trans.page.IsMigrating = true
	mmu.pageTable.Update(trans.page)
	trans.migration = migrationReq
	mmu.isDoingMigration = true
	mmu.currentOnDemandMigration = trans
	mmu.migrationQueue = mmu.migrationQueue[1:]

	return true
}

func (mmu *MMU) markPageAsNotMigratingIfNotInTheMigrationQueue(
	page vm.Page,
) vm.Page {
	inQueue := false
	for _, t := range mmu.migrationQueue {
		if page.PAddr == t.page.PAddr {
			inQueue = true
			break
		}
	}

	if !inQueue {
		page.IsMigrating = false
		mmu.pageTable.Update(page)
		return page
	}

	return page
}

func (mmu *MMU) sendTranlationRsp(
	now sim.VTimeInSec,
	trans transaction,
) (madeProgress bool) {
	req := trans.req
	page := trans.page

	rsp := vm.TranslationRspBuilder{}.
		WithSendTime(now).
		WithSrc(mmu.topPort).
		WithDst(req.Src).
		WithRspTo(req.ID).
		WithPage(page).
		Build()
	mmu.topSender.Send(rsp)

	return true
}

func (mmu *MMU) processMigrationReturn(now sim.VTimeInSec) bool {
	item := mmu.migrationPort.Peek()
	if item == nil {
		return false
	}

	if !mmu.topSender.CanSend(1) {
		return false
	}

	req := mmu.currentOnDemandMigration.req
	page, found := mmu.pageTable.Find(req.PID, req.VAddr)
	if !found {
		panic("page not found")
	}

	rsp := vm.TranslationRspBuilder{}.
		WithSendTime(now).
		WithSrc(mmu.topPort).
		WithDst(req.Src).
		WithRspTo(req.ID).
		WithPage(page).
		Build()
	mmu.topSender.Send(rsp)

	mmu.isDoingMigration = false

	page = mmu.markPageAsNotMigratingIfNotInTheMigrationQueue(page)
	page.IsPinned = true
	mmu.pageTable.Update(page)

	mmu.migrationPort.Retrieve(now)

	return true
}

func (mmu *MMU) parseFromTop(now sim.VTimeInSec) bool {
	maxlength:=0
	for i = 0 ; i < len(mmu.walkingTranslations); i++{
		reqtmp2 := mmu.walkingTranslations[i].req
		//if mmu.walkingTranslations[i].validflg == 0 ||reqtmp2.Remote != 1{
		if reqtmp2.Remote != 1{
			maxlength ++
		}
	}  
	if maxlength >= mmu.maxRequestsInFlight {
		waitptw	+= 1
		return false
	}
	// if len(mmu.walkingTranslations) >= mmu.maxRequestsInFlight {
	// 	waitptw	+= 1
	// 	//fmt.Println("waitptw", waitptw)
	// 	return false
	// }
	

	req := mmu.topPort.Retrieve(now)
	if req == nil {
		return false
	}
	
	tracing.TraceReqReceive(req, now, mmu)
	
	switch req := req.(type) {
	case *vm.TranslationReq:
		mmu.startWalking(req)
	default:
		log.Panicf("MMU canot handle request of type %s", reflect.TypeOf(req))
	}
	
	return true
}

func (mmu *MMU) startWalking(req *vm.TranslationReq) {


	page, found := mmu.pageTable.Find(req.PID, req.VAddr)  //get page 
	found = found
	page = page
	lastaccess = false 
	lastaccesstmp =false
	ioactuallatency = 0
	actuallatency = 0

	hit_in_iommu_pwc_flag := false
	found_in_iommupt_trans_flag := 0
	iolevel_transac := 0

	iommu_cachelevel_latency[-2] = 5*each_pwc_level_latency
	iommu_cachelevel_latency[0] = 4*each_pwc_level_latency
	iommu_cachelevel_latency[1] = 3*each_pwc_level_latency
	iommu_cachelevel_latency[2] = 2*each_pwc_level_latency
	iommu_cachelevel_latency[3] = each_pwc_level_latency

	//implement decimal to binary
	Bin := convertToBin(int(req.VAddr))


	// for len(Bin) < 57{
	// 	Bin = "0" + Bin[0:]
	// }
	// //fmt.Println("req.VAddr", req.VAddr,  Bin)	
 	// slides[0] = Bin[0:9]
	// slides[1] = Bin[0:18]
	// slides[2] = Bin[0:27]
	// slides[3] = Bin[0:36]

	
	for len(Bin) < 35{
  		Bin = "0" + Bin[0:]
 	}
 	//fmt.Println(Bin)
 	//slide_len := 3
 	slides[0] = Bin[0:9]
 	slides[1] = Bin[0:11]
 	slides[2] = Bin[0:19]
 	slides[3] = Bin[0:23]
	

	//fmt.Println("req.VAddr", req.VAddr, slides, Bin)

	cache_level = -2
	otherPWClevel = -2
	hit_loc = -1
	hit_other_loc = -1

	
	//check if PTE is valid in local page table
	if iommutlb.Lastvisitlookup(req.VAddr, req.PID) != -1{
		lastaccess = iommutlb.Validinlocal(req.VAddr, req.DeviceID, req.PID)
	}
	
	//fmt.Println(lastaccess)

	for k = 3; k >= 0; k--{
		for i = 0; i < cache_entry; i++{
			if slides[k] == page_walk_cache[req.DeviceID][i]{
				cache_level = k
				hit_loc = i
				break
			}
		}
		if cache_level != -2{
			break 
		}
	}

	iommu_cache_level = -2
	hit_loc = -1
	for k = 3; k >= 0; k--{
		for i = 0; i < iommu_cache_entry; i++{
			if slides[k] == IOMMU_PWC[i]{
				iommu_cache_level = k
				hit_loc = i
				hit_in_iommu_pwc_flag = true
				break
			}
		}
		if iommu_cache_level != -2{
			break 
		}
	}

	if iommu_cache_level != -2 {
		req_hitin_iommupwc += 1
		eachgpu_hit_io_pwc[req.DeviceID] += 1
	}
	
	iolevel_transac = iommu_cachelevel_latency[iommu_cache_level]

	//fmt.Println("local cache level", cache_level, "io cache level", iommu_cache_level)

	/*============ Baseline ============= */

	foundflag = -1
	mmusetID = int(req.VAddr / setpagesize % uint64(mmuset))
	foundflag = iommutlb.LookupMMUTLB(req.VAddr, mmusetID, req.PID)

	//compute how many cycle need wait in IOMMU PTW, or local PTW

	
	for i = 0; i < len(mmu.walkingTranslations); i++{
		reqtmp := mmu.walkingTranslations[i].req
		mmusetID = int(reqtmp.VAddr / setpagesize % uint64(mmuset))
		

		//only walk GMMU
		if reqtmp.Gmmuhit == 1 {
		 	actuallatency = mmu.walkingTranslations[i].cycleLeft - mmu.walkingTranslations[i].localwait
		 	ioactuallatency = 0
	 	}

	 	//fmt.Println("latency os handle", reqtmp.VAddr, mmu.walkingTranslations[i].os_handle_flag)
		
		// not hit in local, need iommu
		if reqtmp.Gmmuhit != 1 {

			// ioactuallatency = mmu.walkingTranslations[i].cycleLeft  - (2*pcie)  - mmu.walkingTranslations[i].locallevel*2
			// actuallatency = mmu.walkingTranslations[i].cycleLeft - (2*pcie) - mmu.walkingTranslations[i].locallevel - mmu.walkingTranslations[i].iowait - mmu.walkingTranslations[i].iolevel - mmu.walkingTranslations[i].localwait
			

			// hit in iommu tlb
			if mmu.walkingTranslations[i].iommu_tlb_hit_flag != -1 {
				ioactuallatency = mmu.walkingTranslations[i].cycleLeft  - (2*pcie)  - (mmu.walkingTranslations[i].locallevel*2)
			} else {
				// trick io
				ioactuallatency = mmu.walkingTranslations[i].cycleLeft  - (2*pcie)  - (mmu.walkingTranslations[i].locallevel*2) - mmu.walkingTranslations[i].localwait
				// real io
				//ioactuallatency = mmu.walkingTranslations[i].cycleLeft  - (2*pcie)  - (mmu.walkingTranslations[i].locallevel*2) - mmu.walkingTranslations[i].localwait  - mmu.walkingTranslations[i].iowait 
			}


			// if mmu.walkingTranslations[i].direct_os_flag == true {
			// 	actuallatency = mmu.walkingTranslations[i].cycleLeft - (2*pcie) - mmu.walkingTranslations[i].locallevel - mmu.walkingTranslations[i].localwait
			// }else {

			//actuallatency = mmu.walkingTranslations[i].cycleLeft - (2*pcie) - mmu.walkingTranslations[i].locallevel - mmu.walkingTranslations[i].localwait - mmu.walkingTranslations[i].iowait - mmu.walkingTranslations[i].iolevel
			actuallatency = mmu.walkingTranslations[i].cycleLeft - (2*pcie) - mmu.walkingTranslations[i].locallevel - mmu.walkingTranslations[i].localwait - mmu.walkingTranslations[i].iowait - mmu.walkingTranslations[i].iolevel
		
			

			if mmu.walkingTranslations[i].os_handle_flag == true {
				ioactuallatency -= (UVM_BATCH_FIXED_LATENCY + mmu.walkingTranslations[i].oswait)
				actuallatency -= (UVM_BATCH_FIXED_LATENCY + mmu.walkingTranslations[i].oswait)
			}
			
		}
		//fmt.Println("ioactuallatency", ioactuallatency)

		if actuallatency > 0 {
			tmplocalptw[reqtmp.DeviceID] = append(tmplocalptw[reqtmp.DeviceID], actuallatency)
		}

		tmpreq_in_iommutlb := iommutlb.LookupMMUTLB(reqtmp.VAddr, mmusetID, reqtmp.PID)
		//if reqtmp.DeviceID != lastaccesstmp && ioactuallatency > 0 && mmu.walkingTranslations[i].validflg == 0{
		//if iommutlb.Lastvisitlookup(reqtmp.VAddr) != -1 && lastaccesstmp == false && ioactuallatency > 0 && mmu.walkingTranslations[i].validflg == 0{
		
		// miss in tlb, miss in gmmu
		if tmpreq_in_iommutlb == -1 && reqtmp.Gmmuhit != 1 && ioactuallatency > 0 && mmu.walkingTranslations[i].direct_os_flag == false{
		
			tmpiommuptw = append(tmpiommuptw, ioactuallatency)
			//tmplocalptw[reqtmp.DeviceID] = append(tmplocalptw[reqtmp.DeviceID], mmu.walkingTranslations[i].cycleLeft)	
		}
		//fmt.Println("mmu.walkingTranslations[i].cycleLeft2", mmu.walkingTranslations[i].req,mmu.walkingTranslations[i].cycleLeft)
	}


	
	total_io_wait_queue_len += len(tmpiommuptw)
	startWalking_round += 1
	
	

	//fmt.Println(len(tmpiommuptw))
	//iommuPTW: walk thread (16), 
	if len(tmpiommuptw) >= iommuPTW{
		tmpiommu = append(tmpiommuptw[:iommuPTW])
		exist := iommuPTW
		times := len(tmpiommuptw) - iommuPTW 
		
		for k = 0; k < times; k++ {
			sort.Slice(tmpiommu, func(i, j int) bool {return tmpiommu[i] < tmpiommu[j]})
			tmpiommu = append(tmpiommu, tmpiommuptw[exist]+ tmpiommu[k] )
			tmpiommuptw[exist] = tmpiommuptw[exist] - tmpiommu[k]
			exist ++
			//fmt.Println("after  sort tmpiommu",  tmpiommu)
		}

		sort.Slice(tmpiommu, func(i, j int) bool {return tmpiommu[i] < tmpiommu[j]})
		//fmt.Println("finnal sort tmpiommu",  tmpiommu)
		iommuptw_wait = tmpiommu[times]
		//fmt.Println("iommuwait", iommuptw_wait)
	}

	

	for p = 1; p < 9; p++{
		if len(tmplocalptw[p]) >= localPTW{
			//fmt.Println("lenlocalwait", len(tmplocalptw[p]))
			tmplocal[p] = append(tmplocalptw[p][:localPTW])
			exist := localPTW
			times := len(tmplocalptw[p]) - localPTW 
			for k = 0; k < times; k++{
				sort.Slice(tmplocal[p], func(i, j int) bool {return tmplocal[p][i] < tmplocal[p][j]})
				tmplocal[p] = append(tmplocal[p], tmplocalptw[p][exist]+ tmplocal[p][k] )
				tmplocalptw[p][exist] = tmplocalptw[p][exist] - tmplocal[p][k]
				exist ++
			}
			sort.Slice(tmplocal[p], func(i, j int) bool {return tmplocal[p][i] < tmplocal[p][j]})
			localptw_wait[p] = tmplocal[p][times]
			
		}
	}


	/*
	fmt.Println("=================request gpu", req.DeviceID, "req vaddr", req.VAddr)
	if iommutlb.Lastvisitlookup(req.VAddr) == -1 {
		fmt.Println("need insert")
	}else {
		fmt.Println("-page position", iommutlb.GetLastvisit(req.VAddr))
		fmt.Println("-req DeviceID",req.DeviceID, "lastvisit",iommutlb.CheckLastVisit(req.VAddr))
	}
	*/

	mmusetID = int(req.VAddr / setpagesize % uint64(mmuset))

	lastaccess = false
	if iommutlb.Lastvisitlookup(req.VAddr,req.PID) != -1 {
		lastaccess = iommutlb.Validinlocal(req.VAddr, req.DeviceID,req.PID)
		
	}


	wherevisit := iommutlb.Lastvisitlookup(req.VAddr,req.PID) 
	
	//lastvisit里没有记录，说明是cold miss
	if wherevisit == -1 {
		iommutlb.InsertLastVisit(req.VAddr, req.DeviceID,req.PID )
	}

	
	//-----
	//fmt.Println(wherevisit, iommutlb.GetLastvisit(req.VAddr) != req.DeviceID)
	//lastvisit里有记录，所以在某个gpu里
	if wherevisit != -1 &&  iommutlb.GetLastvisit(req.VAddr,req.PID) != req.DeviceID{
		iommutlb.UpdateAccesstime(req.VAddr, req.DeviceID, wherevisit)
		if iommutlb.Checkcounter(req.VAddr, int(req.DeviceID), req.PID) == true{
			
			req.Migrate = 1 //跟req.gmmuhit对得上
			iommutlb.UpdateLastvisit(req.VAddr, req.DeviceID,req.PID)
			

			//test_tmp += 1
		}
	}



	// find in which gpu
	req_gpu_id := req.DeviceID-1
	page_pos_gpu_id := -1
	for g := 0; g < 8; g++{
		gmmu_page, found_in_gmmu := GmmuPageTable[g].Find(req.PID, req.VAddr)
		gmmu_page = gmmu_page
		if found_in_gmmu == true {
			page_pos_gpu_id = g 
			break
		}
	}

	//fmt.Println("vaddr", req.VAddr,"page pos", page_pos_gpu_id, "pid", req.PID)

	//fmt.Println("-page position gmmu pt", page_pos_gpu_id+1, req.VAddr)
	hit_in_gmmu_flag := false
	// not in any GPU, insert
	if page_pos_gpu_id == -1 {
		gmmu_page := page 
		GmmuPageTable[req_gpu_id].Insert(gmmu_page)

		//req_gointo_iommu += 1
		eachgpu_compulsory_miss[req.DeviceID] += 1
		total_compulsory_miss += 1

	} else if uint64(page_pos_gpu_id) == uint64(req_gpu_id) { 
		// gmmu hit
		gmmu_page, found_in_gmmu := GmmuPageTable[req_gpu_id].Find(req.PID, req.VAddr)
		found_in_gmmu = found_in_gmmu
		GmmuPageTable[req_gpu_id].Update(gmmu_page)

		req_hitin_gmmu += 1
		eachgpu_hit_gmmu[req.DeviceID] += 1

		hit_in_gmmu_flag = true
		req.Gmmuhit = 1
		iolevel_transac = 0


	} else if uint64(page_pos_gpu_id) != uint64(req_gpu_id) {
		//in other gpu, need migrate, update gmmu page table
		page_need_migrate, found_in_gpu := GmmuPageTable[page_pos_gpu_id].Find(req.PID, req.VAddr)
		found_in_gpu = found_in_gpu
		GmmuPageTable[page_pos_gpu_id].Remove(req.PID, req.VAddr)
		GmmuPageTable[req_gpu_id].Insert(page_need_migrate)

		//flush l2 tlb
		iommutlb.DeleteL2TLB(page_pos_gpu_id, req.VAddr, req.PID)
		// iommutlb记录里有，需要flush
		if iommutlb.LookupMMUTLB(req.VAddr, mmusetID, req.PID) != -1{
			iommutlb.FlushMMUTLB(req.VAddr, req.PID)
		}

		eachgpu_inothergpu_miss[req.DeviceID] += 1
		total_inothergpu_miss += 1
		//req_gointo_iommu += 1
		//test_tmp2 += 1
	}

	
	
	if req.Migrate == 1{
		needmigrateflg = 1
		for l := 1; l < 5; l++{
			if iommutlb.Lastvisitlookup(req.VAddr,req.PID) != -1 && iommutlb.Validinlocal(req.VAddr, uint64(l),req.PID) == false{
				unnecessary[l]++
			}
		}
	}else{
		needmigrateflg = 0
	}
	/*
	fmt.Println("+page position", iommutlb.GetLastvisit(req.VAddr))
	fmt.Println("+req DeviceID",req.DeviceID, "lastvisit",iommutlb.CheckLastVisit(req.VAddr))
	fmt.Println("+", req.Migrate, "<Migrate, Gmmuhit>", req.Gmmuhit, lastaccess, "=======================", "total req.Migrate/gmmu migrate", test_tmp, test_tmp2)
	*/

	
	//	gmmu miss && iommutlb miss 才要计算pwc latency
	found_in_iommutlb0 := -1
	if hit_in_gmmu_flag == false{
		found_in_iommutlb0 = iommutlb.LookupMMUTLB(req.VAddr, mmusetID, req.PID)
	}


	locallatency, mmu.latency = mmu.setlatencybaseline(req, cache_level, iommu_cache_level, found_in_iommutlb0, hit_in_gmmu_flag, req.DeviceID, iommuptw_wait, localptw_wait[req.DeviceID])
	

	
	// check if need add remote access latency
	remoteaccess := 0
	

	if req.Gmmuhit != 1 {
		mmu.latency += pcie
	}


	direct_goto_os_flag := false
	// cold miss or first access
	key := addrNpid{req.PID, req.VAddr}
	if request_record[key] == 0 {
		direct_goto_os_flag = true
		request_record[key] = 1
	} else { // hot miss
		request_record[key] += 1
		direct_goto_os_flag = false
	}

	// --------------------------------------------------------
	if ENABLE_DIRECT_OS == false {
		direct_goto_os_flag = false
	}


	tmp_os_flag = false
	if hit_in_gmmu_flag == false{
		found_in_iommutlb := iommutlb.LookupMMUTLB(req.VAddr, mmusetID, req.PID)

		
		if direct_goto_os_flag == true { // directly go to os handling
			// add to iommu pt
			iommuPageTable.Insert(page)

			mmu.latency += UVM_BATCH_FIXED_LATENCY
			tmp_os_flag = true
			req_gointo_batch += 1
			eachgpu_goto_os[req.DeviceID] += 1
			eachgpu_oshandle_latency[req.DeviceID] += UVM_BATCH_FIXED_LATENCY
			fault_num += 1

			saved_latency_by_directos = saved_latency_by_directos + iommu_cachelevel_latency[iommu_cache_level] + iommuptw_wait
			mmu.latency = mmu.latency - iommu_cachelevel_latency[iommu_cache_level] - iommuptw_wait
			iommuptw_wait_total -= iommuptw_wait
			eachgpu_io_wait_latency[req.DeviceID] -= iommuptw_wait
			iommuptw_wait = 0
			iolevel_transac = 0
			
			if iommu_cache_level != -2{
				req_hitin_iommupwc -= 1
				eachgpu_hit_io_pwc[req.DeviceID] -= 1
			}
			iommu_pwc_latency -= iommu_cachelevel_latency[iommu_cache_level]
			eachgpu_io_pwc_latency[req.DeviceID] -= iommu_cachelevel_latency[iommu_cache_level]

		} else { // goto iommu should = hit in iommu
			req_gointo_iommu += 1
			eachgpu_goto_iommu[req.DeviceID] += 1

			if found_in_iommutlb == -1{

				iommu_page, found_in_iommupt := iommuPageTable.Find(req.PID, req.VAddr) 
				//false: not in iommu page tble
				if found_in_iommupt == false{
					fmt.Println("if use direct_goto_os_flag, this message should not appear.")
					iommu_page = page 
					iommuPageTable.Insert(iommu_page)

					// find the translation lantency
					mmu.latency += UVM_BATCH_FIXED_LATENCY
					tmp_os_flag = true
					req_gointo_batch += 1
					eachgpu_goto_os[req.DeviceID] += 1
					eachgpu_oshandle_latency[req.DeviceID] += UVM_BATCH_FIXED_LATENCY
					fault_num += 1

				}else {
					req_hitin_iommupt += 1
					eachgpu_hit_io_pt[req.DeviceID] += 1
					req_hitin_iommu += 1
					eachgpu_hit_io_total[req.DeviceID] += 1

					if hit_in_iommu_pwc_flag == true {
						req_hitin_iommupwc_and_pt += 1
					}
					tmp_os_flag = false
					iommuPageTable.Update(iommu_page)
					found_in_iommupt_trans_flag = 1
				}
			} else{
				tmp_os_flag = false
				req_hitin_iommutlb += 1
				eachgpu_hit_io_tlb[req.DeviceID] += 1
				req_hitin_iommu += 1
				eachgpu_hit_io_total[req.DeviceID] += 1
				iolevel_transac = 0
			}
		}
			
		
	}
	

	// add os wait latency
	os_wait := 0 
	if tmp_os_flag == true {
		os_wait = mmu.calculateHostWaitLatency(hostThread)
		mmu.latency += os_wait

		ultimate_os_wait += os_wait
		eachgpu_os_wait_latency[req.DeviceID] += os_wait
	}



	ultimate_total_latency += mmu.latency
	eachgpu_alllatency[req.DeviceID] += mmu.latency
	ultimate_total_requests += 1
	eachgpu_total_requests[req.DeviceID] += 1


	
	fmt.Println("v2.0---iommu batch fixed latency:", UVM_BATCH_FIXED_LATENCY, "batchThreshold:", batchThreshold, "direct goto os flag", direct_goto_os_flag)
	/*
	if batch_num != 0 {
		fmt.Println("total waiting cycle of batch", total_waiting_time_of_batch, ", ave", total_waiting_time_of_batch/batch_num, ", batch num",batch_num)
	} else {
		fmt.Println("total waiting cycle of batch", total_waiting_time_of_batch, "batch_num = 0")
	} 
	*/
	fmt.Println("total_requests:", ultimate_total_requests, ", hit gmmu:", req_hitin_gmmu, ", go into iommu:", req_gointo_iommu ,", goto os:",req_gointo_batch, "all compulsory", total_compulsory_miss, "all other gpu miss", total_inothergpu_miss, "ave io wait queue", total_io_wait_queue_len/startWalking_round)
	fmt.Println("iommu_tlb:", req_hitin_iommutlb, ", pwc:", req_hitin_iommupwc, "pwc&pt", req_hitin_iommupwc_and_pt, ", pt:", req_hitin_iommupt, "hit total:", req_hitin_iommu, "iopwc hit rate", io_hitrate)
	fmt.Println("LATENCY breakdown: local wait:",localwaittotal, ", local memory access:", local_memory_access,", total local latency:", totallatency,", interconnect:", pcieall,
			", iommuptw total wait:", iommuptw_wait_total, "iommu memory access" , iommu_pwc_latency ,"oswait", ultimate_os_wait,", os handle:", req_gointo_batch*UVM_BATCH_FIXED_LATENCY, ", replay", replayall ,", all latency:", ultimate_total_latency)
	//fmt.Println("saved latency by direct os:", saved_latency_by_directos)

	fmt.Println("++each_total_requests", eachgpu_total_requests, "gmmu_hit", eachgpu_hit_gmmu, "goto iommu", eachgpu_goto_iommu, "goto os" , eachgpu_goto_os, "compulsory", eachgpu_compulsory_miss, "in other gpu", eachgpu_inothergpu_miss)
	fmt.Println("++each_iommu_tlb", eachgpu_hit_io_tlb, "pwc", eachgpu_hit_io_pwc, "pt", eachgpu_hit_io_pt, "io total hit", eachgpu_hit_io_total)
	fmt.Println("++each_latency_local_wait", eachgpu_local_wait_latency, "local pwc", eachgpu_local_pwc_latency, "interconnect", eachgpu_pcieall_latency, "io wait",  eachgpu_io_wait_latency, "io pwc", eachgpu_io_pwc_latency, 
			"os_wait", eachgpu_os_wait_latency, "os handle", eachgpu_oshandle_latency, "replay", eachgpu_replayall_latency, "all latency", eachgpu_alllatency)

	/*============ End Baseline ============= */
 	if direct_goto_os_flag == true {
 		iolevel_transac = 0
 		iommuptw_wait = 0
 	}
	
	translationInPipeline := transaction{
		req:       req,
		cycleLeft: mmu.latency,
		local:   needmigrateflg,  
		locallevel:     locallatency,
		localwait: localptw_wait[int(req.DeviceID)],
		iowait: iommuptw_wait,
		oswait: os_wait,
		iolevel: iolevel_transac,
		validflg: 0,
		validid: int(req.DeviceID),
		remote: remoteaccess,
		iommu_tlb_hit_flag: found_in_iommutlb0,
		os_handle_flag: tmp_os_flag,
		direct_os_flag: direct_goto_os_flag,
		iommupt_hit_flag: found_in_iommupt_trans_flag,
	}
	if req.Remote != 1 {
		requestPTW[int(req.DeviceID)] ++
	}
	
	localwaitlatency += localptw_wait[int(req.DeviceID)] 
	mmu.walkingTranslations = append(mmu.walkingTranslations, translationInPipeline)
	//fmt.Println("translationInPipeline1\n", mmu.walkingTranslations)
	//fmt.Println("len(mmu.walkingTranslations)", len(mmu.walkingTranslations))


	for i = 0; i < 9; i++{
		tmplocalptw[i] = tmplocalptw[i][:0]
		tmpiommuptw = tmpiommuptw[:0]
		tmplocal[i] = tmplocal[i][:0]
		tmpiommu = tmpiommu[:0]
		localptw_wait[i] = 0
		iommuptw_wait = 0
	}

	
	
}

func (mmu *MMU) toRemove(index int) bool {
	for i := 0; i < len(mmu.toRemoveFromPTW); i++ {
		remove := mmu.toRemoveFromPTW[i]
		if remove == index {
			return true
		}
	}
	return false
}

func unique(intSlice []uint64) []uint64 {
	keys := make(map[int]bool)
	list := []uint64{}
	for _, entry := range intSlice {
		if _, value := keys[int(entry)]; !value {
			keys[int(entry)] = true
			list = append(list, entry)
		}
	}
	return list
}

func (mmu *MMU) calculateHostWaitLatency (hostThread int) int{
	os_req_num := 0
	cal_i := 0

	if (len(mmu.walkingTranslations) < hostThread) {
		return 0
	}
	
	for cal_i = 0; cal_i < len(mmu.walkingTranslations); cal_i++{
		reqtmp := mmu.walkingTranslations[cal_i].req
		if mmu.walkingTranslations[cal_i].os_handle_flag == true {
			actual_host_latency := 0

			if reqtmp.Gmmuhit == 1  || mmu.walkingTranslations[cal_i].iommu_tlb_hit_flag != -1 || mmu.walkingTranslations[cal_i].iommupt_hit_flag  == 1{
				actual_host_latency = 0
			} else {
				// if mmu.walkingTranslations[cal_i].direct_os_flag == true  {
				// 	// real
				// 	// actual_host_latency = mmu.walkingTranslations[cal_i].cycleLeft - (pcie*2) - mmu.walkingTranslations[cal_i].locallevel*2 - mmu.walkingTranslations[cal_i].localwait - mmu.walkingTranslations[cal_i].oswait
				
				// 	actual_host_latency = mmu.walkingTranslations[cal_i].cycleLeft - (pcie) - mmu.walkingTranslations[cal_i].locallevel - mmu.walkingTranslations[cal_i].oswait - mmu.walkingTranslations[cal_i].iowait
				// }else {
				
				//real 
				//actual_host_latency = mmu.walkingTranslations[cal_i].cycleLeft - (2*pcie) - mmu.walkingTranslations[cal_i].locallevel*2 - mmu.walkingTranslations[cal_i].localwait - mmu.walkingTranslations[cal_i].iowait - mmu.walkingTranslations[cal_i].iolevel - mmu.walkingTranslations[cal_i].oswait
				//trick
				actual_host_latency = mmu.walkingTranslations[cal_i].cycleLeft - (pcie)  - mmu.walkingTranslations[cal_i].iowait - mmu.walkingTranslations[cal_i].oswait
			
				
			}
			if actual_host_latency > 0{
				os_req_num += 1
			}
		}
	}
	// trick
	oswait := UVM_BATCH_FIXED_LATENCY * (os_req_num/hostThread)
	//real os
	//oswait := UVM_BATCH_FIXED_LATENCY * int64((os_req_num-1)/hostThread)

	fmt.Println("+++==+++++os_req_num", os_req_num, "oswait", oswait)
	if oswait > 0 {
		return oswait
	} else {
		oswait = 0
	}
	

	// oswait := cal_0

	return oswait
}


// func (mmu *MMU) calculateHostWaitLatency (hostThread int) int{
// 	os_req_num := 0
// 	cal_i := 0
// 	oswait := 0
// 	var tmposptw []int
// 	var	tmpos []int
// 	for cal_i = 0; cal_i < len(mmu.walkingTranslations); cal_i++{
// 		reqtmp := mmu.walkingTranslations[cal_i].req
// 		if mmu.walkingTranslations[cal_i].os_handle_flag == true {
// 			actual_host_latency := 0

// 			if reqtmp.Gmmuhit == 1  || mmu.walkingTranslations[cal_i].iommu_tlb_hit_flag != -1 || mmu.walkingTranslations[cal_i].iommupt_hit_flag  == 1{
// 				actual_host_latency = 0
// 			} else {
// 				if mmu.walkingTranslations[cal_i].direct_os_flag == true  {
// 					actual_host_latency = mmu.walkingTranslations[cal_i].cycleLeft - (2*pcie) - mmu.walkingTranslations[cal_i].locallevel*2 - mmu.walkingTranslations[cal_i].localwait - mmu.walkingTranslations[cal_i].oswait
// 				}else {
// 					actual_host_latency = mmu.walkingTranslations[cal_i].cycleLeft - (2*pcie) - mmu.walkingTranslations[cal_i].locallevel*2 - mmu.walkingTranslations[cal_i].localwait - mmu.walkingTranslations[cal_i].iowait - mmu.walkingTranslations[cal_i].iolevel - mmu.walkingTranslations[cal_i].oswait
// 				}
				
// 			}
// 			if actual_host_latency > 0{
// 				os_req_num += 1
// 				tmposptw = append(tmposptw, actual_host_latency)
// 			}
// 		}
// 	}

// 	if len(tmposptw) >= hostThread{
// 		tmpos = append(tmposptw[:hostThread])
// 		exist := hostThread
// 		times := len(tmposptw) - hostThread

// 		for cal_k := 0; cal_k < times; cal_k++{
// 			sort.Slice(tmpos , func(i, j int) bool {return tmpos[i] < tmpos[j]})
// 			tmpos = append(tmpos, tmposptw[exist]+ tmpos[cal_k] )
// 			tmposptw[exist] = tmposptw[exist] - tmpos[cal_k]
// 			exist ++
// 		}
// 		sort.Slice(tmpos, func(i, j int) bool {return tmpos[i] < tmpos[j]})
// 		oswait = tmpos[times]
// 	}

// 	fmt.Println("+++==+++++os_req_num", os_req_num, "oswait", oswait)


// 	return oswait
// }

func (mmu *MMU) setlatencybaseline (req *vm.TranslationReq, cache_level int, iommu_cache_level int, found_in_iommutlb0 int, hit_in_gmmu_flag bool, ReqID uint64, iommuptw_wait int, localptw_wait int)(locallatency int, latency int){
	
	if cache_level == 3{
		locallatency = each_pwc_level_latency 
	}else if cache_level == 2{
		locallatency =  2*each_pwc_level_latency
	}else if cache_level == 1{
		locallatency = 3*each_pwc_level_latency 
	}else if cache_level == 0{
		locallatency = 4*each_pwc_level_latency 
	}else if cache_level == -2{
		locallatency = 5*each_pwc_level_latency
		
	}

	local_memory_access += locallatency
	eachgpu_local_pwc_latency[req.DeviceID] += locallatency
	//locallatency = 5*each_pwc_level_latency
	if hit_in_gmmu_flag == false {
		pcieall = pcieall + pcie*2
		eachgpu_pcieall_latency[req.DeviceID] = eachgpu_pcieall_latency[req.DeviceID] + pcie*2
		replayall = replayall + locallatency
		eachgpu_replayall_latency[req.DeviceID] = eachgpu_replayall_latency[req.DeviceID] + locallatency
	}
	
	if hit_in_gmmu_flag == true{
		mmu.latency = locallatency + localptw_wait
	}else if found_in_iommutlb0 == -1 && hit_in_gmmu_flag != true{
		if iommu_cache_level == -2{
			mmu.latency = pcie + each_pwc_level_latency*5 + locallatency*2 + localptw_wait + iommuptw_wait //pcie + iommu_ptw + replay + local_ptw
			io_hitrate[4]++

			iommu_only_latency += each_pwc_level_latency*5 + iommuptw_wait 
			iommu_pwc_latency += 5*each_pwc_level_latency
			eachgpu_io_pwc_latency[req.DeviceID] += 5*each_pwc_level_latency
		}else if iommu_cache_level == 0{
			mmu.latency = pcie + each_pwc_level_latency*4 + locallatency*2 + localptw_wait + iommuptw_wait
			io_hitrate[0]++

			iommu_only_latency += each_pwc_level_latency*4 + iommuptw_wait
			iommu_pwc_latency += 4*each_pwc_level_latency
			eachgpu_io_pwc_latency[req.DeviceID] += 4*each_pwc_level_latency
		}else if iommu_cache_level == 1{
			mmu.latency = pcie + each_pwc_level_latency*3 + locallatency*2 + localptw_wait + iommuptw_wait
			io_hitrate[1]++

			iommu_only_latency += each_pwc_level_latency*3 + iommuptw_wait
			iommu_pwc_latency += 3*each_pwc_level_latency
			eachgpu_io_pwc_latency[req.DeviceID] += 3*each_pwc_level_latency
		}else if iommu_cache_level == 2{
			mmu.latency = pcie + each_pwc_level_latency*2 + locallatency*2 + localptw_wait + iommuptw_wait
			io_hitrate[2]++

			iommu_only_latency += each_pwc_level_latency*2 + iommuptw_wait 
			iommu_pwc_latency += 2*each_pwc_level_latency
			eachgpu_io_pwc_latency[req.DeviceID] += 2*each_pwc_level_latency
		}else if iommu_cache_level == 3{
			mmu.latency = pcie + each_pwc_level_latency + locallatency*2 + localptw_wait + iommuptw_wait				
			io_hitrate[3]++

			iommu_only_latency += each_pwc_level_latency + iommuptw_wait 
			iommu_pwc_latency += each_pwc_level_latency
			eachgpu_io_pwc_latency[req.DeviceID] += each_pwc_level_latency
		}
		
		
		
		
		iommuptw_wait_total += iommuptw_wait
		eachgpu_io_wait_latency[req.DeviceID] += iommuptw_wait
		iototallatency += int64(mmu.latency)
		pflocallatency += int64(locallatency + localptw_wait) 
		resolvelatency += int64(locallatency)
	} else if found_in_iommutlb0 != -1 && hit_in_gmmu_flag != true {
		mmu.latency = pcie + locallatency*2 + localptw_wait
	}
	//fmt.Println(" ============== ",locallatency, localptw_wait)
	totallatency = totallatency + int64(locallatency + localptw_wait)
	localwaittotal += localptw_wait
	eachgpu_local_wait_latency[req.DeviceID] += localptw_wait
	
	return locallatency, mmu.latency
}

func (mmu *MMU) setlatencyideal (cache_level int, iommu_cache_level int, foundflag int, lastaccess bool, ReqID uint64, iommuptw_wait int, localptw_wait int)(locallatency int, latency int){
	
	if cache_level == 3{
		locallatency = each_pwc_level_latency 
	}else if cache_level == 2{
		locallatency =  2*each_pwc_level_latency
	}else if cache_level == 1{
		locallatency = 3*each_pwc_level_latency 
	}else if cache_level == 0{
		locallatency = 4*each_pwc_level_latency 
	}else if cache_level == -2{
		locallatency = 5*each_pwc_level_latency
	}
	
	mmu.latency = locallatency + localptw_wait
	
	return locallatency, mmu.latency
}
