package um

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	var tests = []struct {
		Input  []byte
		Result []uint32
	}{
		{[]byte{0x00, 0x00, 0x00, 0x00}, []uint32{0x00000000}},
		{[]byte{0xde, 0xad, 0xca, 0xfe}, []uint32{0xdeadcafe}},
		{[]byte{0xab, 0xcd, 0xef, 0x00, 0x01, 0x02, 0x03, 0x04}, []uint32{0xabcdef00, 0x01020304}},
	}

	for _, test := range tests {
		got := Convert(test.Input)
		if !reflect.DeepEqual(got, test.Result) {
			t.Errorf("expected '%v' but got '%v'", test.Result, got)
		}
	}
}

func TestOp(t *testing.T) {
	testData := map[uint32]uint32{
		0b00000000_00000000_00000000_00000000: 0,
		0b00010000_00000000_00000000_00000000: 1,
		0b10100000_00000000_00000000_00000000: 10,
		0b11010000_00000000_00000000_00000000: 13,
	}

	for instr, expected := range testData {
		got := Op(instr)
		if got != expected {
			t.Errorf("with '%x' expected '%d' but got '%d'", instr, expected, got)
		}
	}
}

func TestReg(t *testing.T) {
	var tests = []struct {
		Func     func(uint32) int
		Instr    uint32
		Expected int
	}{
		{RegA, 0b000000000, 0},
		{RegA, 0b001000000, 1},
		{RegA, 0b111000000, 7},
		{RegB, 0b000000, 0},
		{RegB, 0b001000, 1},
		{RegB, 0b111000, 7},
		{RegC, 0b000, 0},
		{RegC, 0b001, 1},
		{RegC, 0b111, 7},
	}

	for _, test := range tests {
		got := test.Func(test.Instr)
		if got != test.Expected {
			t.Errorf("with '%x' expected '%d' but got '%d'", test.Instr, test.Expected, got)
		}
	}
}

func TestCondMove(t *testing.T) {
	expected := uint32(0xf0f0f0f0)
	UM.Registers[0] = 1
	UM.Registers[1] = expected
	// "random" value
	UM.Registers[2] = (expected + 1) & 0xffffffff
	instr := uint32(0b010_001_000)
	OpCondMove(instr)
	got := UM.Registers[2]

	if got != expected {
		t.Errorf("with '%b' expected '%x' but got '%x'", instr, expected, got)
	}
}

func TestArrayIdx(t *testing.T) {
	instr := uint32(0b000_001_000)
	array := alloc(4)
	offset := uint32(2)
	expected := uint32(0xfaceface)
	UM.Registers[1] = array
	UM.Registers[0] = offset
	UM.Platters[array] = Platter{0x0, 0x0, expected, 0x0}
	OpArrayIdx(instr)
	got := UM.Registers[0]
	if got != expected {
		t.Errorf("expected '%x' but got '%x'", expected, got)
	}

	abandonAll()
}

func TestArrayAmd(t *testing.T) {
	instr := uint32(0b010_001_000)
	array := alloc(4)
	offset := uint32(1)
	expected := uint32(0xbeefcafe)
	UM.Registers[2] = array
	UM.Registers[1] = offset
	UM.Registers[0] = expected
	OpArrayAmd(instr)
	got := UM.Platters[array][offset]
	if got != expected {
		t.Errorf("expected '%x' but got '%x'", expected, got)
	}
	abandonAll()
}

func TestMath(t *testing.T) {
	var tests = []struct {
		Func     func(uint32)
		B        uint32
		C        uint32
		Expected uint32
	}{
		{OpAdd, 0x0000, 0x0000, 0x0000},
		{OpAdd, 0x0000, 0x0001, 0x0001},
		{OpAdd, 0x0001, 0x0001, 0x0002},
		{OpAdd, 0xff00, 0x00ff, 0xffff},
		{OpAdd, 0xffffffff, 0x00000001, 0x00000000},
		{OpMult, 0x0000, 0x0000, 0x0000},
		{OpMult, 0x0001, 0x0000, 0x0000},
		{OpMult, 0x0001, 0x0001, 0x0001},
		{OpMult, 0xffff, 0xffff, 0xfffe0001},
		{OpMult, 0xffffffff, 0xffffffff, 0x0001},
		{OpDiv, 0x0000, 0x0001, 0x0000},
		{OpDiv, 0x0100, 0x0010, 0x0010},
		{OpDiv, 0xffff, 0xffff, 0x0001},
		{OpDiv, 0xfedcba98, 0x01234567, 0x00e0},
		{OpDiv, 0xffffffff, 0xffffffff, 0x00000001},
		{OpNand, 0x0000, 0x0000, 0xffffffff},
		{OpNand, 0xffeeddcc, 0x76543210, 0x89bbefff},
		{OpNand, 0x01234567, 0xfedcba98, 0xffffffff},
	}

	instr := uint32(0b000_001_010)
	for _, test := range tests {
		// Fill with "random" value
		UM.Registers[0] = uint32((test.Expected + 1) & 0xffffffff)
		UM.Registers[1] = test.B
		UM.Registers[2] = test.C
		test.Func(instr)
		got := UM.Registers[0]
		if got != test.Expected {
			t.Errorf("with '%x' and '%x' expected '%x' but got '%x'", test.B, test.C, test.Expected, got)
		}
	}
}

func TestHalt(t *testing.T) {
	UM.Halt = false
	OpHalt(uint32(0))
	if !UM.Halt {
		t.Errorf("expected true for UM.Halt, got false")
	}
	UM.Halt = false
}

func abandonAll() {
	UM.Platters = []Platter{{}}
}

// TODO more tests
func TestAlloc(t *testing.T) {
	instr := uint32(0b001_000)
	l := len(UM.Platters)
	size := uint32(10)
	UM.Registers[0] = size
	OpAlloc(instr)

	if len(UM.Platters) != l+1 {
		t.Errorf("after allocation the number of platters should increase by 1. before: '%d', after: '%d'", l, len(UM.Platters))
	}

	if UM.Registers[1] == 0 {
		t.Errorf("allocated platter number must not be 0")
	}

	abandonAll()
}

func alloc(size int) uint32 {
	instr := uint32(0b001_000)
	UM.Registers[0] = uint32(size)
	OpAlloc(instr)

	return UM.Registers[1]
}

func TestAbandon(t *testing.T) {
	var arrays []uint32
	instr := uint32(0b001_000)

	for i := 0; i < 10; i++ {
		arrays = append(arrays, alloc(1000))
	}

	l := len(UM.Platters)
	if l != 11 {
		t.Errorf("After 10 allocations there should be 11 active platters got '%d'", l)
	}

	for _, id := range arrays {
		UM.Registers[0] = id
		OpAbandon(instr)
		l = len(UM.Platters[id])
		if l != 0 {
			t.Errorf("After freeing up a platter its length should be 0 got '%d'", l)
		}
	}
}

func TestOutWorker(t *testing.T) {
	instr := uint32(0b000)
	chr := 'x'
	UM.Registers[0] = uint32(chr)
	buf := bytes.NewBufferString("")
	outWorker(instr, buf)
	expected := string(chr)
	got := fmt.Sprintf("%s", buf)
	if got != expected {
		t.Errorf("expected '%s' but got '%s'", expected, got)
	}
}

func TestInWorker(t *testing.T) {
	inWorker(uint32(0), strings.NewReader("a"))
	if UM.Registers[0] != uint32('a') {
		t.Errorf("expected 'a' but got '%c'", UM.Registers[0])
	}
}

func TestRegSpec(t *testing.T) {
	tests := map[uint32]int{
		0b000_0000000000000000000000000: 0,
		0b001_0000000000000000000000000: 1,
		0b111_0000000000000000000000000: 7,
	}

	for instr, expected := range tests {
		got := RegSpec(instr)
		if got != expected {
			t.Errorf("expected '%d' but got '%d'", expected, got)
		}
	}
}

func TestValueSpec(t *testing.T) {
	tests := map[uint32]uint32{
		0b000_0000000000000000000000000: 0,
		0b000_0000000000000000000000001: 1,
		//		0b000_0000000000000000000000000: 0,
		0b000_1000000000000000000000000: 0x1000000,
		0b000_1111111111111111111111111: 0x1ffffff,
	}

	for instr, expected := range tests {
		got := ValueSpec(instr)
		if got != expected {
			t.Errorf("expected '%x' but got '%x'", expected, got)
		}
	}
}

// TODO test if platter copy is a real copy not a reference
func TestLoad(t *testing.T) {
	instr := uint32(0b001_000)
	array := alloc(4)
	prg := Platter{0x0, 0x0, 0x0, 0xdeaddead}
	finger := uint32(3)
	UM.Platters[array] = prg
	UM.Registers[1] = array
	UM.Registers[0] = finger
	OpLoad(instr)
	if !reflect.DeepEqual(UM.Platters[0], prg) {
		t.Errorf("platter 0 is not identical to source platter: '%v' '%v'", UM.Platters[0], prg)
	}

	got := UM.Finger
	if got != finger {
		t.Errorf("finger is expected to point to '%x' but points to '%x'", finger, got)
	}
	abandonAll()
}

func TestOrtho(t *testing.T) {
	tests := []uint32{
		0x0000,
		0x0001,
		0x00ff,
		0xff00,
		0xffff,
		0xffffff,
		0x1ffffff,
	}

	for _, test := range tests {
		instr := uint32(7 << 25)
		instr += test
		// Fill with "random" value
		UM.Registers[7] = uint32((test + 1) & 0xffffffff)
		OpOrtho(instr)
		got := UM.Registers[7]
		if got != test {
			t.Errorf("expected '%x' but got '%x'", test, got)
		}
	}
}

func benchmarkAlloc(size uint32, b *testing.B) {
	for n := 0; n < b.N; n++ {
		instr := uint32(0b000)
		UM.Registers[0] = size
		OpAlloc(instr)
	}
}

func BenchmarkAlloc1k(b *testing.B) {
	benchmarkAlloc(1_000, b)
}

func BenchmarkAlloc10k(b *testing.B) {
	benchmarkAlloc(10_000, b)
}

func BenchmarkAlloc100k(b *testing.B) {
	benchmarkAlloc(100_000, b)
}

func BenchmarkAlloc1M(b *testing.B) {
	benchmarkAlloc(1_000_000, b)
}

func BenchmarkAlloc10M(b *testing.B) {
	benchmarkAlloc(10_000_000, b)
}

func BenchmarkAlloc100M(b *testing.B) {
	benchmarkAlloc(100_000_000, b)
}
