package main

type J interface {
	Method()
}

type (
	U16  uint16
	U32  uint32
	U64  uint64
	U128 [2]uint64
	F32  float32
	F64  float64
	C128 complex128
	S    string
	B    []byte
	M    map[int]int
	C    chan int
	Z    struct{}
)

func (U16) Method()  {}
func (U32) Method()  {}
func (U64) Method()  {}
func (U128) Method() {}
func (F32) Method()  {}
func (F64) Method()  {}
func (C128) Method() {}
func (S) Method()    {}
func (B) Method()    {}
func (M) Method()    {}
func (C) Method()    {}
func (Z) Method()    {}

var (
	u16  = U16(1)
	u32  = U32(2)
	u64  = U64(3)
	u128 = U128{4, 5}
	f32  = F32(6)
	f64  = F64(7)
	c128 = C128(8 + 9i)
	s    = S("10")
	b    = B("11")
	m    = M{12: 13}
	c    = make(C, 14)
	z    = Z{}
	p    = &z
	pp   = &p
)

var (
	iu16  interface{} = u16
	iu32  interface{} = u32
	iu64  interface{} = u64
	iu128 interface{} = u128
	if32  interface{} = f32
	if64  interface{} = f64
	ic128 interface{} = c128
	is    interface{} = s
	ib    interface{} = b
	im    interface{} = m
	ic    interface{} = c
	iz    interface{} = z
	ip    interface{} = p
	ipp   interface{} = pp

	ju16  J = u16
	ju32  J = u32
	ju64  J = u64
	ju128 J = u128
	jf32  J = f32
	jf64  J = f64
	jc128 J = c128
	js    J = s
	jb    J = b
	jm    J = m
	jc    J = c
	jz J = z
	jp J = p // The method set for *T contains the methods for T.
	// pp does not implement error.
)

func second(a ...interface{}) interface{} {
	return a[1]
}

func main() {
	
}