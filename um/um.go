package um

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

type Registers [8]uint32

type Platter []uint32

type Machine struct {
	Registers Registers
	Platters  []Platter
	Finger    uint32
	Halt      bool
}

var UM = Machine{Platters: []Platter{{}}}

var Ops = []func(uint32){
	OpCondMove,
	OpArrayIdx,
	OpArrayAmd,
	OpAdd,
	OpMult,
	OpDiv,
	OpNand,
	OpHalt,
	OpAlloc,
	OpAbandon,
	OpOut,
	OpIn,
	OpLoad,
	OpOrtho,
}

func ReadProgram(prg string) []byte {
	f, err := os.Open(prg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()

	stats, statsErr := f.Stat()
	if statsErr != nil {
		panic(statsErr)
	}
	var size int64 = stats.Size()
	b := make([]byte, size)

	n, err := f.Read(b)
	fmt.Println("Read", n, "bytes")

	return b
}

func Convert(scroll []byte) []uint32 {
	var res []uint32
	for idx := 0; idx < len(scroll); idx += 4 {
		a := uint32(scroll[idx]) << 24
		b := uint32(scroll[idx+1]) << 16
		c := uint32(scroll[idx+2]) << 8
		d := uint32(scroll[idx+3])
		res = append(res, uint32(a+b+c+d))
	}

	return res
}

func Init(prg string) {
	UM.Platters[0] = Convert(ReadProgram(prg))
}

func Spin() {
	var opCnt [14]int

	for {
		txt := fmt.Sprintf("Finger: %d", UM.Finger)
		UMScr.Registers.MvAddStr(0, 0, txt)
		instr := UM.Platters[0][UM.Finger]
		op := Op(instr)
		opCnt[op] += 1
		UM.Finger += 1
		Ops[op](instr)
		if UM.Halt {
			fmt.Println(opCnt)
			return
		}
	}
}

func Op(instr uint32) uint32 {
	return instr >> 28
}

func RegA(instr uint32) int {
	return int(instr&0b111000000) >> 6
}

func RegB(instr uint32) int {
	return int(instr&0b111000) >> 3
}

func RegC(instr uint32) int {
	return int(instr & 0b111)
}

func RegSpec(instr uint32) int {
	return int(instr>>25) & 0b111
}

func ValueSpec(instr uint32) uint32 {
	return instr & 0x1ffffff
}

func OpCondMove(instr uint32) {
	if UM.Registers[RegC(instr)] != 0 {
		UM.Registers[RegA(instr)] = UM.Registers[RegB(instr)]
	}
}

func OpArrayIdx(instr uint32) {
	array := UM.Registers[RegB(instr)]
	offset := UM.Registers[RegC(instr)]
	UM.Registers[RegA(instr)] = UM.Platters[array][offset]
}

func OpArrayAmd(instr uint32) {
	array := UM.Registers[RegA(instr)]
	offset := UM.Registers[RegB(instr)]
	UM.Platters[array][offset] = UM.Registers[RegC(instr)]
}

func OpAdd(instr uint32) {
	res := (UM.Registers[RegB(instr)] + UM.Registers[RegC(instr)]) & 0xffffffff
	UM.Registers[RegA(instr)] = res
}

func OpMult(instr uint32) {
	res := (UM.Registers[RegB(instr)] * UM.Registers[RegC(instr)]) & 0xffffffff
	UM.Registers[RegA(instr)] = res
}

func OpDiv(instr uint32) {
	res := (UM.Registers[RegB(instr)] / UM.Registers[RegC(instr)]) & 0xffffffff
	UM.Registers[RegA(instr)] = res
}

func OpNand(instr uint32) {
	res := UM.Registers[RegB(instr)] & UM.Registers[RegC(instr)]
	UM.Registers[RegA(instr)] = ^res
}

func OpHalt(instr uint32) {
	UM.Halt = true
}

func OpAlloc(instr uint32) {
	size := UM.Registers[RegC(instr)]
	UM.Platters = append(UM.Platters, make(Platter, size))
	UM.Registers[RegB(instr)] = uint32(len(UM.Platters) - 1)
}

func OpAbandon(instr uint32) {
	array := int(UM.Registers[RegC(instr)])
	UM.Platters[array] = Platter{}
}

func OpOut(instr uint32) {
	outWorker(instr, os.Stdout)
}

func outWorker(instr uint32, w io.Writer) {
	chr := UM.Registers[RegC(instr)]
	fmt.Fprintf(w, "%c", chr)
}

func OpIn(instr uint32) {
	inWorker(instr, bufio.NewReader(os.Stdin))
}

func inWorker(instr uint32, reader io.Reader) {
	buf := make([]byte, 1)
	reader.Read(buf)

	UM.Registers[RegC(instr)] = uint32(buf[0])
}

func OpLoad(instr uint32) {
	array := UM.Registers[RegB(instr)]
	cpy := make(Platter, len(UM.Platters[array]))
	copy(cpy, UM.Platters[array])
	UM.Platters[0] = cpy
	UM.Finger = UM.Registers[RegC(instr)]
}

func OpOrtho(instr uint32) {
	UM.Registers[RegSpec(instr)] = ValueSpec(instr)
}
