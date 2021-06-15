package msgpack

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"unsafe"
)

// Type represents the type of value in the MessagePack stream.
type Type int

// list of MessagePack types.
const (
	Invalid Type = iota
	Nil
	Bool
	Int
	Uint
	Float
	ArrayLen
	MapLen
	String
	Binary
	Extension
)

var typeNames = [...]string{
	Invalid:   "Invalid",
	Nil:       "Nil",
	Bool:      "Bool",
	Int:       "Int",
	Uint:      "Uint",
	Float:     "Float",
	ArrayLen:  "ArrayLen",
	MapLen:    "MapLen",
	String:    "String",
	Binary:    "Binary",
	Extension: "Extension",
}

// String returns a string representation of the Type.
func (t Type) String() string {
	var n string

	if 0 <= t && t < Type(len(typeNames)) {
		n = typeNames[t]
	}
	if n == "" {
		n = "unknown"
	}

	return n
}

// ErrDataSizeTooLarge is the data size too large error.
var ErrDataSizeTooLarge = errors.New("msgpack: data size too large")

// ExtensionMap specifies functions for converting MessagePack extensions to Go
// values.
//
// The key is the MessagePack extension type.
// The value is a function that converts the extension data to a Go value.
type ExtensionMap map[int]func([]byte) (interface{}, error)

// Decoder reads MessagePack objects from an io.Reader.
type Decoder struct {
	extensions ExtensionMap
	err        error
	r          *bufio.Reader
	n          uint64
	p          []byte
	t          Type
	peek       bool
}

const bufioReaderSize = 4096

// NewDecoder allocates and initializes a new decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: bufio.NewReaderSize(r, bufioReaderSize),
	}
}

// SetExtensions specifies functions for converting MessagePack extensions to Go
// values.
func (d *Decoder) SetExtensions(extensions ExtensionMap) {
	d.extensions = extensions
}

// Type returns the type of the current value in the stream.
func (d *Decoder) Type() Type {
	return d.t
}

// Extension returns the type of the current Extension value.
func (d *Decoder) Extension() int {
	return int(d.n)
}

// Bytes returns the current String, Binary or Extension value as a slice of
// bytes.
func (d *Decoder) Bytes() []byte {
	if d.peek {
		p := make([]byte, len(d.p))
		copy(p, d.p)
		d.p = p
	}

	return d.p
}

// BytesNoCopy returns the current String, Binary or Extension value as a slice
// of bytes. The underlying array may point to data that will be overwritten by
// a subsequent call to Unpack.
func (d *Decoder) BytesNoCopy() []byte {
	return d.p
}

// String returns the current String, Binary or Extension value as a string.
func (d *Decoder) String() string {
	return *(*string)(unsafe.Pointer(&d.p))
}

// Int returns the current Int value.
func (d *Decoder) Int() int64 {
	return int64(d.n)
}

// Uint returns the current Uint value.
func (d *Decoder) Uint() uint64 {
	return d.n
}

// Len returns the current ArrayLen or MapLen value.
func (d *Decoder) Len() int {
	return int(d.n)
}

// Bool returns the current Bool value.
func (d *Decoder) Bool() bool {
	if d.n != 0 {
		return true
	}

	return false
}

// Float returns the current Float value.
func (d *Decoder) Float() float64 {
	return math.Float64frombits(d.n)
}

// Unpack reads the next value from the MessagePack stream. Call Type to get the
// type of the current value. Call Bool, Uint, Int, Float, Bytes or Extension
// to get the value.
func (d *Decoder) Unpack() error {
	if d.err != nil {
		return d.err
	}

	code, err := d.r.ReadByte()
	if err != nil {
		// Don't call d.fatal here because we don't want io.EOF converted to
		// io.ErrUnexpectedEOF
		d.err = err
		return err
	}

	f := formats[code]
	d.t = f.t

	d.n, err = f.fn(d, code)
	if err != nil {
		return d.fatal(err)
	}

	if !f.more {
		d.p = nil
		return nil
	}

	nn := int(d.n)
	if nn < 0 {
		return d.fatal(ErrDataSizeTooLarge)
	}

	if f.t == Extension {
		var b byte
		b, err = d.r.ReadByte()
		if err != nil {
			return d.fatal(err)
		}
		d.n = uint64(b)
	}

	if nn <= bufioReaderSize {
		d.peek = true
		d.p, err = d.r.Peek(nn)
		if err != nil {
			return d.fatal(err)
		}
		d.r.Discard(nn)
	} else {
		d.peek = false
		d.p = make([]byte, nn)
		_, err := io.ReadFull(d.r, d.p)
		if err != nil {
			return d.fatal(err)
		}
	}

	return nil
}

// Skip skips over any nested values in the stream.
func (d *Decoder) Skip() error {
	n := d.skipCount()

	for n > 0 {
		n--
		if err := d.Unpack(); err != nil {
			return err
		}
		n += d.skipCount()
	}

	return nil
}

func (d *Decoder) skipCount() int {
	switch d.Type() {
	case ArrayLen:
		return d.Len()
	case MapLen:
		return 2 * d.Len()
	default:
		return 0
	}
}

var formats = [256]*struct {
	t    Type
	fn   func(d *Decoder, code byte) (uint64, error)
	more bool
}{
	fixIntCodeMin: {
		t:  Int,
		fn: func(d *Decoder, code byte) (uint64, error) { return uint64(code), nil },
	},
	fixMapCodeMin: {
		t:  MapLen,
		fn: func(d *Decoder, code byte) (uint64, error) { return uint64(code) - uint64(fixMapCodeMin), nil },
	},
	fixArrayCodeMin: {
		t:  ArrayLen,
		fn: func(d *Decoder, code byte) (uint64, error) { return uint64(code) - uint64(fixArrayCodeMin), nil },
	},
	fixStringCodeMin: {
		t:    String,
		fn:   func(d *Decoder, code byte) (uint64, error) { return uint64(code) - uint64(fixStringCodeMin), nil },
		more: true,
	},
	nilCode: {
		t:  Nil,
		fn: func(d *Decoder, code byte) (uint64, error) { return 0, nil },
	},
	unusedCode: {
		t: Invalid,
		fn: func(d *Decoder, code byte) (uint64, error) {
			return 0, fmt.Errorf("msgpack: unknown format code %x", code)
		},
	},
	falseCode: {
		t:  Bool,
		fn: func(d *Decoder, code byte) (uint64, error) { return 0, nil },
	},
	trueCode: {
		t:  Bool,
		fn: func(d *Decoder, code byte) (uint64, error) { return 1, nil },
	},
	binary8Code: {
		t:    Binary,
		fn:   (*Decoder).read1,
		more: true,
	},
	binary16Code: {
		t:    Binary,
		fn:   (*Decoder).read2,
		more: true,
	},
	binary32Code: {
		t:    Binary,
		fn:   (*Decoder).read4,
		more: true,
	},
	ext8Code: {
		t:    Extension,
		fn:   (*Decoder).read1,
		more: true,
	},
	ext16Code: {
		t:    Extension,
		fn:   (*Decoder).read2,
		more: true,
	},
	ext32Code: {
		t:    Extension,
		fn:   (*Decoder).read4,
		more: true,
	},
	float32Code: {
		t: Float,
		fn: func(d *Decoder, code byte) (uint64, error) {
			n, err := d.read4(code)
			return math.Float64bits(float64(math.Float32frombits(uint32(n)))), err
		},
	},
	float64Code: {
		t:  Float,
		fn: (*Decoder).read8,
	},
	uint8Code: {
		t:  Uint,
		fn: (*Decoder).read1,
	},
	uint16Code: {
		t:  Uint,
		fn: (*Decoder).read2,
	},
	uint32Code: {
		t:  Uint,
		fn: (*Decoder).read4,
	},
	uint64Code: {
		t:  Uint,
		fn: (*Decoder).read8,
	},
	int8Code: {
		t: Int,
		fn: func(d *Decoder, code byte) (uint64, error) {
			n, err := d.read1(code)
			return uint64(int64(int8(n))), err
		},
	},
	int16Code: {
		t: Int,
		fn: func(d *Decoder, code byte) (uint64, error) {
			n, err := d.read2(code)
			return uint64(int64(int16(n))), err
		},
	},
	int32Code: {
		t: Int,
		fn: func(d *Decoder, code byte) (uint64, error) {
			n, err := d.read4(code)
			return uint64(int64(int32(n))), err
		},
	},
	int64Code: {
		t:  Int,
		fn: (*Decoder).read8,
	},
	fixExt1Code: {
		t:    Extension,
		fn:   func(d *Decoder, code byte) (uint64, error) { return 1, nil },
		more: true,
	},
	fixExt2Code: {
		t:    Extension,
		fn:   func(d *Decoder, code byte) (uint64, error) { return 2, nil },
		more: true,
	},
	fixExt4Code: {
		t:    Extension,
		fn:   func(d *Decoder, code byte) (uint64, error) { return 4, nil },
		more: true,
	},
	fixExt8Code: {
		t:    Extension,
		fn:   func(d *Decoder, code byte) (uint64, error) { return 8, nil },
		more: true,
	},
	fixExt16Code: {
		t:    Extension,
		fn:   func(d *Decoder, code byte) (uint64, error) { return 16, nil },
		more: true,
	},
	string8Code: {
		t:    String,
		fn:   (*Decoder).read1,
		more: true,
	},
	string16Code: {
		t:    String,
		fn:   (*Decoder).read2,
		more: true,
	},
	string32Code: {
		t:    String,
		fn:   (*Decoder).read4,
		more: true,
	},
	array16Code: {
		t:  ArrayLen,
		fn: (*Decoder).read2,
	},
	array32Code: {
		t:  ArrayLen,
		fn: (*Decoder).read4,
	},
	map16Code: {
		t:  MapLen,
		fn: (*Decoder).read2,
	},
	map32Code: {
		t:  MapLen,
		fn: (*Decoder).read4,
	},
	negFixIntCodeMin: {
		t:  Int,
		fn: func(d *Decoder, code byte) (uint64, error) { return uint64(int64(int8(code))), nil },
	},
}

func init() {
	for i := fixIntCodeMin + 1; i <= fixIntCodeMax; i++ {
		formats[i] = formats[fixIntCodeMin]
	}

	for i := fixMapCodeMin + 1; i <= fixMapCodeMax; i++ {
		formats[i] = formats[fixMapCodeMin]
	}

	for i := fixArrayCodeMin + 1; i <= fixArrayCodeMax; i++ {
		formats[i] = formats[fixArrayCodeMin]
	}

	for i := fixStringCodeMin + 1; i <= fixStringCodeMax; i++ {
		formats[i] = formats[fixStringCodeMin]
	}

	for i := negFixIntCodeMin + 1; i <= negFixIntCodeMax; i++ {
		formats[i] = formats[negFixIntCodeMin]
	}
}

func (d *Decoder) fatal(err error) error {
	if err == io.EOF {
		err = io.ErrUnexpectedEOF
	}

	d.t = Invalid
	d.err = err
	return err
}

func (d *Decoder) read1(byte) (uint64, error) {
	b, err := d.r.ReadByte()

	return uint64(b), err
}

func (d *Decoder) read2(byte) (uint64, error) {
	p, err := d.r.Peek(2)
	if err != nil {
		return 0, err
	}
	d.r.Discard(2)

	return uint64(p[1]) | uint64(p[0])<<8, nil
}

func (d *Decoder) read4(byte) (uint64, error) {
	p, err := d.r.Peek(4)
	if err != nil {
		return 0, err
	}
	d.r.Discard(4)

	return uint64(p[3]) | uint64(p[2])<<8 | uint64(p[1])<<16 | uint64(p[0])<<24, nil
}

func (d *Decoder) read8(byte) (uint64, error) {
	p, err := d.r.Peek(8)
	if err != nil {
		return 0, err
	}
	d.r.Discard(8)

	return uint64(p[7]) | uint64(p[6])<<8 | uint64(p[5])<<16 | uint64(p[4])<<24 |
		uint64(p[3])<<32 | uint64(p[2])<<40 | uint64(p[1])<<48 | uint64(p[0])<<56, nil
}
