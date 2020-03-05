package main

import(
	"emulator"
	"iop_dma"
)

type Voice struct {
	left_vol, right_vol uint16
    pitch uint16
    adsr1, adsr2 uint16
    current_envelope uint16

    start_addr uint32
    current_addr uint32
    loop_addr uint32
    loop_addr_specified bool

    counter uint32
    block_pos int
	loop_code int
}

func (v *Voice) reset() {
	v.left_vol = 0
	v.right_vol = 0
	v.pitch = 0
	v.adsr1 = 0
	v.adsr2 = 0
	v.current_envelope = 0
	v.start_addr = 0
	v.current_addr = 0
	v.loop_addr = 0
	v.loop_addr_specified = false
	v.counter = 0
	v.block_pos = 0
	v.loop_code = 0
}

type SPU_STAT struct {
	DMA_finished bool
    DMA_busy bool
}

var id int