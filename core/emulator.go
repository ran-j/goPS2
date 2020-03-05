package main

import(
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "bufio"
    "bytes"
    "github.com/utahta/go-atomicbool"
	"ps2/lib/memcpy"
	"ps2/lib/cbind"
    "iop/iop_dma"
    "iop/Cop0"
    "iop/Cop1"
    "iop/cdvd"
    "iop/iop"
    "iop/iop_dma"
    "iop/iop_timers"
    "iop/gamepad"
    "iop/memcard"
    "iop/sio2"
    "iop/spu"
    "ee/intc"
    "ee/ipu/ipu"
    "ee/dmac"
    "ee/emotion"
    "ee/timers"
    "ee/vif"
    "ee/vu"
    "gs"
    "gif"
    "scheduler"
    "sif"
)

type SKIP_HACK_STRUCT struct {  
    NONE int
    LOAD_ELF int
    LOAD_DISC int
}

type CPU_MODE_STRUCT struct {  
    DONT_CARE int
    JIT int
    INTERPRETER int
}

var SKIP_HACK = SKIP_HACK_STRUCT{ 0,1,2 }
var CPU_MODE = CPU_MODE_STRUCT{ 0,1,2 } 


var save_requested bool
var load_requested bool
var gsdump_requested bool
var gsdump_single_frame bool
var gsdump_running bool
// var save_requested = atomicbool.New(false)
// var load_requested = atomicbool.New(false)
// var gsdump_requested = atomicbool.New(false)
// var gsdump_single_frame = atomicbool.New(false)
// var gsdump_running = atomicbool.New(false)
var save_state_path string
var frames int
var cp0 Cop0
var cp0 fpu
var cdvd CDVD_Drive
var dmac DMAC
var cpu EmotionEngine
var timers EmotionTiming
var pad Gamepad
var gs GraphicsSynthesizer
var gif GraphicsInterface
var iop IOP
var iop_dma IOP_DMA
var iop_timers IOPTiming
var intc INTC
var ipu ImageProcessingUnit
var memcard Memcard
var scheduler Scheduler
var sio2 SIO2
var spu, spu2 SPU
var sif SubsystemInterface
var vif0, vif1 VectorInterface
var vu0, vu1 VectorUnit
var VBLANK_sent bool
var cop2_interlock, vu_interlock bool

var ee_log bytes.Buffer // is that correct ??
var ee_stdout string

// vu1_run_func(v *VectorUnit, i int)

var RDRAM uint64
var IOP_RAM *uint8
var BIOS *uint8
var SPU_RAM *uint8

var scratchpad = make([]uint8, 1024 * 16)
var iop_scratchpad = make([]uint8, 1024)
 
var iop_scratchpad_start uint32
var MCH_RICM, MCH_DRD uint32
var rdram_sdevid uint8

var IOP_POST uint8
var IOP_I_STAT uint32
var IOP_I_MASK uint32
var IOP_I_CTRL uint32

var skip_BIOS_hack = SKIP_HACK

var ELF_file *uint8
var ELF_size *uint32

var frame_ended bool

type emulator struct { }

func (e emulator) run() {
    gs.start_frame()
    VBLANK_sent = false
    originalRounding := Fegetround()
    // Fegetround(FE_TOWARDZERO) ??
    if (save_requested)
        // save_state(save_state_path)
    if (load_requested)
        // load_state(save_state_path)
    if (gsdump_requested)
    {
        gsdump_requested = false
        gs.send_dump_request()
        gsdump_running = !gsdump_running
    }
    else if (gsdump_single_frame)
    {
        gs.send_dump_request()
        if (gsdump_running)
        {
            gsdump_running = false
            gsdump_single_frame = false
        }
        else
        {
            gsdump_running = true
        }
    }

    frame_ended = false

    // add_ee_event(VBLANK_START, &Emulator::vblank_start, VBLANK_START_CYCLES)
    // add_ee_event(VBLANK_END, &Emulator::vblank_end, CYCLES_PER_FRAME)

    for frame_ended != true {
        ee_cycles := scheduler.calculate_run_cycles()
        bus_cycles := scheduler.get_bus_run_cycles()
        iop_cycles := scheduler.get_iop_run_cycles()
        scheduler.update_cycle_counts()

        cpu.run(ee_cycles)
        iop_timers.run(iop_cycles)
        iop_dma.run(iop_cycles)
        iop.run(iop_cycles)
        iop.interrupt_check(IOP_I_CTRL && (IOP_I_MASK & IOP_I_STAT))

        dmac.run(bus_cycles)
        timers.run(bus_cycles)
        ipu.run()
        vif0.update(bus_cycles)
        vif1.update(bus_cycles)
        gif.run(bus_cycles)

        //VU's run at EE speed, however VU0 maintains its own speed
        vu0.run(ee_cycles)
        vu1_run_func(vu1, ee_cycles)

        scheduler.process_events(this)
	}
}

func (e emulator) reset() {
    save_requested = false
    load_requested = false
    gsdump_requested = false
    iop_i_ctrl_delay = 0
    ee_stdout = ""
    frames = 0
    skip_BIOS_hack = skip_BIOS_hack.NONE
    if (!RDRAM)
        RDRAM = uint64(1024 * 1024 * 32)
    if (!IOP_RAM)
        IOP_RAM = uint64(1024 * 1024 * 2)  
    if (!BIOS)
        BIOS = uint64(1024 * 1024 * 4) 
    if (!SPU_RAM)
        SPU_RAM = uint64(1024 * 1024 * 2)
    
    cdvd.reset()
    cp0.reset()
    cp0.init_mem_pointers(RDRAM, BIOS, uint8(&scratchpad))
    cpu.reset()
    cpu.init_tlb()
    dmac.reset(RDRAM, uint8(&scratchpad))
    fpu.reset()
    gs.reset()
    gif.reset()
    iop.reset()
    iop_dma.reset(IOP_RAM)
    iop_timers.reset()
    intc.reset()
    ipu.reset()
    pad.reset()
    scheduler.reset()
    sif.reset()
    sio2.reset()
    spu.reset(SPU_RAM)
    spu2.reset(SPU_RAM)
    timers.reset()
    vif0.reset()
    vif1.reset()
    vu0.reset()
    vu1.reset()    
}

func (e *emulator) load_BIOS(BIOS_file *uint32) {
    if (!BIOS)
		BIOS := [...]uint32{1024 * 1024 * 4}

	shm.Memcpy(BIOS, BIOS_file, 1024 * 1024 * 4)
}

func (e *emulator) load_ELF(ELF *uint32, size uint32) {
    if (ELF[0] != 0x7F || ELF[1] != 'E' || ELF[2] != 'L' || ELF[3] != 'F')
    {
        fmt.Println("Invalid elf\n");
        return;
    }
    fmt.Println("Valid elf\n");
    ELF_file = uint32(size);
    ELF_size = size;
    shm.Memcpy(ELF_file, ELF, size);
}

func Emulator() emulator {
    this := emulator
    cdvd(this, &iop_dma),
    cp0(&dmac),
    cpu(&cp0, &fpu, this, &vu0, &vu1),
    dmac(&cpu, this, &gif, &ipu, &sif, &vif0, &vif1, &vu0, &vu1),
    gif(&gs, &dmac),
    gs(&intc),
    iop(this),
    iop_dma(this, &cdvd, &sif, &sio2, &spu, &spu2),
    iop_timers(this),
    intc(this, &cpu),
    ipu(&intc, &dmac),
    timers(&intc),
    sio2(this, &pad, &memcard),
    spu(1, this, &iop_dma),
    spu2(2, this, &iop_dma),
    vif0(nullptr, &vu0, &intc, &dmac, 0),
    vif1(&gif, &vu1, &intc, &dmac, 1),
    vu0(0, this, &intc, &cpu, &vu1),
    vu1(1, this, &intc, &cpu, &vu0),
    sif(&iop_dma, &dmac)
    return this
}
