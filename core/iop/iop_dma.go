package main

import(
	"emulator"
	"spu"
)

var RAM uint64 = uint64(0)

type dma_copy_func func()

type IOP_DMA_Chan_Control struct {
    direction_from bool
	unk8  bool
	sync_mode uint64
	busy bool
	unk30 bool
}

type IOP_DMA_Channel struct {
	addr uint64
	word_count uint64
	size uint64
	block_size uint64
	control IOP_DMA_Chan_Control
	tag_addr uint64
	tag_end bool
	funct dma_copy_func
	dma_req bool
	delay int
	index int
}

type DMA_DPCR struct {
    priorities []uint64
	enable []bool
}

type DMA_DICR struct {
	force_IRQ []bool
    STAT []uint64
	MASK []uint64
	master_int_enable []bool
}

//Merge of DxCR, DxCR2, DxCR3 for easier processing
var DPCR DMA_DPCR = DMA_DPCR{ make([]uint64, 16), make([]bool, 16) }
var DICR DMA_DICR = DMA_DICR{ make([]bool, 2), make([]uint64, 2), make([]uint64, 2), make([]bool, 2) }

var active_channel *IOP_DMA_Channel
var channels [16]IOP_DMA_Channel
 
func CHAN(index int) string {
	borp := [...]string{"MDECin", "MDECout", "GPU", "CDVD", "SPU", "PIO", "OTC", "67", "SPU2", "8", "SIF0", "SIF1", "SIO2in", "SIO2out"}
	return borp[index];
}

func reset(uint64 RAM) {
	RAM = RAM
	active_channel = false

}