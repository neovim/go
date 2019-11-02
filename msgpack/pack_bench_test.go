package msgpack

import (
	"bytes"
	"testing"
)

func BenchmarkPackInt(b *testing.B) {
	benchs := []struct {
		name string
		v    int64
	}{
		{name: "int64(0x0)", v: int64(0x0)},
		{name: "int64(0x1)", v: int64(0x1)},
		{name: "int64(0x7f)", v: int64(0x7f)},
		{name: "int64(0x80)", v: int64(0x80)},
		{name: "int64(0x7fff)", v: int64(0x7fff)},
		{name: "int64(0x8000)", v: int64(0x8000)},
		{name: "int64(0x7fffffff)", v: int64(0x7fffffff)},
		{name: "int64(0x80000000)", v: int64(0x80000000)},
		{name: "int64(0x7fffffffffffffff)", v: int64(0x7fffffffffffffff)},
		{name: "int64(-0x1)", v: int64(-0x1)},
		{name: "int64(-0x20)", v: int64(-0x20)},
		{name: "int64(-0x21)", v: int64(-0x21)},
		{name: "int64(-0x80)", v: int64(-0x80)},
		{name: "int64(-0x81)", v: int64(-0x81)},
		{name: "int64(-0x8000)", v: int64(-0x8000)},
		{name: "int64(-0x8001)", v: int64(-0x8001)},
		{name: "int64(-0x80000000)", v: int64(-0x80000000)},
		{name: "int64(-0x80000001)", v: int64(-0x80000001)},
		{name: "int64(-0x8000000000000000)", v: int64(-0x8000000000000000)},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackInt(bb.v)
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackUint(b *testing.B) {
	benchs := []struct {
		name string
		v    uint64
	}{
		{name: "uint64(0x0)", v: uint64(0x0)},
		{name: "uint64(0x1)", v: uint64(0x1)},
		{name: "uint64(0x7f)", v: uint64(0x7f)},
		{name: "uint64(0xff)", v: uint64(0xff)},
		{name: "uint64(0x100)", v: uint64(0x100)},
		{name: "uint64(0xffff)", v: uint64(0xffff)},
		{name: "uint64(0x10000)", v: uint64(0x10000)},
		{name: "uint64(0xffffffff)", v: uint64(0xffffffff)},
		{name: "uint64(0x100000000)", v: uint64(0x100000000)},
		{name: "uint64(0xffffffffffffffff)", v: uint64(0xffffffffffffffff)},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackUint(bb.v)
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackBool(b *testing.B) {
	benchs := []struct {
		name string
		v    bool
	}{
		{name: "true", v: true},
		{name: "false", v: false},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackBool(bb.v)
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackFloat(b *testing.B) {
	benchs := []struct {
		name string
		v    float64
	}{
		{name: "float64(1.23456)", v: float64(1.23456)},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackFloat(bb.v)
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackArrayLen(b *testing.B) {
	benchs := []struct {
		name string
		v    arrayLen
	}{
		{name: "arrayLen(0x0)", v: arrayLen(0x0)},
		{name: "arrayLen(0x1)", v: arrayLen(0x1)},
		{name: "arrayLen(0xf)", v: arrayLen(0xf)},
		{name: "arrayLen(0x10)", v: arrayLen(0x10)},
		{name: "arrayLen(0xffff)", v: arrayLen(0xffff)},
		{name: "arrayLen(0x10000)", v: arrayLen(0x10000)},
		{name: "arrayLen(0xffffffff)", v: arrayLen(0xffffffff)},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackArrayLen(int64(bb.v))
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackMapLen(b *testing.B) {
	benchs := []struct {
		name string
		v    mapLen
	}{
		{name: "mapLen(0x0)", v: mapLen(0x0)},
		{name: "mapLen(0x1)", v: mapLen(0x1)},
		{name: "mapLen(0xf)", v: mapLen(0xf)},
		{name: "mapLen(0x10)", v: mapLen(0x10)},
		{name: "mapLen(0xffff)", v: mapLen(0xffff)},
		{name: "mapLen(0x10000)", v: mapLen(0x10000)},
		{name: "mapLen(0xffffffff)", v: mapLen(0xffffffff)},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackMapLen(int64(bb.v))
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackString(b *testing.B) {
	benchs := []struct {
		name string
		v    string
	}{
		{name: "string(1234567890123456789012345678901)", v: "1234567890123456789012345678901"},
		{name: "string(12345678901234567890123456789012)", v: "12345678901234567890123456789012"},
		{name: "emptyString", v: ""},
		{name: "1", v: "1"},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackString(bb.v)
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackBinary(b *testing.B) {
	benchs := []struct {
		name string
		v    []byte
	}{
		{name: "[]byte(``)", v: []byte("")},
		{name: "[]byte(`1`)", v: []byte("1")},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackBinary(bb.v)
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackExtension(b *testing.B) {
	benchs := []struct {
		name string
		v    extension
	}{
		{name: "extension{1, ``}", v: extension{1, ""}},
		{name: "extension{2, `1`}", v: extension{2, "1"}},
		{name: "extension{3, `12`}", v: extension{3, "12"}},
		{name: "extension{4, `1234`}", v: extension{4, "1234"}},
		{name: "extension{5, `12345678`}", v: extension{5, "12345678"}},
		{name: "extension{6, `1234567890123456`}", v: extension{6, "1234567890123456"}},
		{name: "extension{7, `12345678901234567`}", v: extension{7, "12345678901234567"}},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackExtension(bb.v.k, []byte(bb.v.d))
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

func BenchmarkPackNil(b *testing.B) {
	benchs := []struct {
		name string
		v    interface{}
	}{
		{name: "nil", v: nil},
	}

	for _, bb := range benchs {
		b.Run(bb.name, func(b *testing.B) {
			var buf bytes.Buffer
			enc := NewEncoder(&buf)
			b.ReportAllocs()
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_ = enc.PackNil()
			}

			b.SetBytes(int64(buf.Len()))
		})
	}
}

// var benchPackTests = []struct {
// 	name string
// 	v    interface{}
// }{
// 	{name: "int64(0x0)", v: int64(0x0)},
// 	{name: "int64(0x1)", v: int64(0x1)},
// 	{name: "int64(0x7f)", v: int64(0x7f)},
// 	{name: "int64(0x80)", v: int64(0x80)},
// 	{name: "int64(0x7fff)", v: int64(0x7fff)},
// 	{name: "int64(0x8000)", v: int64(0x8000)},
// 	{name: "int64(0x7fffffff)", v: int64(0x7fffffff)},
// 	{name: "int64(0x80000000)", v: int64(0x80000000)},
// 	{name: "int64(0x7fffffffffffffff)", v: int64(0x7fffffffffffffff)},
// 	{name: "int64(-0x1)", v: int64(-0x1)},
// 	{name: "int64(-0x20)", v: int64(-0x20)},
// 	{name: "int64(-0x21)", v: int64(-0x21)},
// 	{name: "int64(-0x80)", v: int64(-0x80)},
// 	{name: "int64(-0x81)", v: int64(-0x81)},
// 	{name: "int64(-0x8000)", v: int64(-0x8000)},
// 	{name: "int64(-0x8001)", v: int64(-0x8001)},
// 	{name: "int64(-0x80000000)", v: int64(-0x80000000)},
// 	{name: "int64(-0x80000001)", v: int64(-0x80000001)},
// 	{name: "int64(-0x8000000000000000)", v: int64(-0x8000000000000000)},
// 	{name: "uint64(0x0)", v: uint64(0x0)},
// 	{name: "uint64(0x1)", v: uint64(0x1)},
// 	{name: "uint64(0x7f)", v: uint64(0x7f)},
// 	{name: "uint64(0xff)", v: uint64(0xff)},
// 	{name: "uint64(0x100)", v: uint64(0x100)},
// 	{name: "uint64(0xffff)", v: uint64(0xffff)},
// 	{name: "uint64(0x10000)", v: uint64(0x10000)},
// 	{name: "uint64(0xffffffff)", v: uint64(0xffffffff)},
// 	{name: "uint64(0x100000000)", v: uint64(0x100000000)},
// 	{name: "uint64(0xffffffffffffffff)", v: uint64(0xffffffffffffffff)},
// 	{name: "true", v: true},
// 	{name: "false", v: false},
// 	{name: "float64(1.23456)", v: float64(1.23456)},
// 	{name: "arrayLen(0x0)", v: arrayLen(0x0)},
// 	{name: "arrayLen(0x1)", v: arrayLen(0x1)},
// 	{name: "arrayLen(0xf)", v: arrayLen(0xf)},
// 	{name: "arrayLen(0x10)", v: arrayLen(0x10)},
// 	{name: "arrayLen(0xffff)", v: arrayLen(0xffff)},
// 	{name: "arrayLen(0x10000)", v: arrayLen(0x10000)},
// 	{name: "arrayLen(0xffffffff)", v: arrayLen(0xffffffff)},
// 	{name: "mapLen(0x0)", v: mapLen(0x0)},
// 	{name: "mapLen(0x1)", v: mapLen(0x1)},
// 	{name: "mapLen(0xf)", v: mapLen(0xf)},
// 	{name: "mapLen(0x10)", v: mapLen(0x10)},
// 	{name: "mapLen(0xffff)", v: mapLen(0xffff)},
// 	{name: "mapLen(0x10000)", v: mapLen(0x10000)},
// 	{name: "mapLen(0xffffffff)", v: mapLen(0xffffffff)},
// 	{name: "string(1234567890123456789012345678901)", v: "1234567890123456789012345678901"},
// 	{name: "string(12345678901234567890123456789012)", v: "12345678901234567890123456789012"},
// 	{name: "emptyString", v: ""},
// 	{name: "string(1)", v: "1"},
// 	{name: "[]byte(``)", v: []byte("")},
// 	{name: "[]byte(`1`)", v: []byte("1")},
// 	{name: "extension{1, ``}", v: extension{1, ""}},
// 	{name: "extension{2, `1`}", v: extension{2, "1"}},
// 	{name: "extension{3, `12`}", v: extension{3, "12"}},
// 	{name: "extension{4, `1234`}", v: extension{4, "1234"}},
// 	{name: "extension{5, `12345678`}", v: extension{5, "12345678"}},
// 	{name: "extension{6, `1234567890123456`}", v: extension{6, "1234567890123456"}},
// 	{name: "extension{7, `12345678901234567`}", v: extension{7, "12345678901234567"}},
// 	{name: "nil", v: nil},
// }
//
// func BenchmarkPack(b *testing.B) {
// 	for _, tt := range benchPackTests {
// 		b.Run(tt.name, func(b *testing.B) {
// 			var buf bytes.Buffer
// 			enc := NewEncoder(&buf)
// 			b.ReportAllocs()
// 			b.ResetTimer()
//
// 			// Go Type     Encoder method
// 			// ---------   --------------------
// 			// int64       PackInt
// 			// uint64      PackUint
// 			// bool        PackBool
// 			// float64     PackFloat
// 			// arrayLen    PackArrayLen
// 			// mapLen      PackMapLen
// 			// string      PackString(s, false)
// 			// []byte      PackBytes(s, true)
// 			// extension   PackExtension(k, d)
// 			// nil         PackNil
// 			// --------------------------------
// 			for i := 0; i < b.N; i++ {
// 				var err error
// 				switch v := tt.v.(type) {
// 				case int64:
// 					err = enc.PackInt(v)
// 				case uint64:
// 					err = enc.PackUint(v)
// 				case bool:
// 					err = enc.PackBool(v)
// 				case float64:
// 					err = enc.PackFloat(v)
// 				case arrayLen:
// 					err = enc.PackArrayLen(int64(v))
// 				case mapLen:
// 					err = enc.PackMapLen(int64(v))
// 				case string:
// 					err = enc.PackString(v)
// 				case []byte:
// 					err = enc.PackBinary(v)
// 				case extension:
// 					err = enc.PackExtension(v.k, []byte(v.d))
// 				case nil:
// 					err = enc.PackNil()
// 				default:
// 					err = fmt.Errorf("no pack for type %T", v)
// 				}
// 				if err != nil {
// 					b.Fatal(err)
// 				}
//
// 				_ = buf.Bytes()
// 			}
//
// 			b.SetBytes(int64(buf.Len()))
// 		})
// 	}
// }
