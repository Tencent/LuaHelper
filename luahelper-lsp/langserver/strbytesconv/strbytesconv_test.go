package strbytesconv

import (
	"reflect"
	"testing"
)

func Test_StringToBytes(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name      string
		args      args
		wantBytes []byte
	}{
		// TODO: Add test cases.
		{
			name: "testStrToBytes",
			args: args{
				str: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
			},
			wantBytes: []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBytes := StringToBytes(tt.args.str)
			if !reflect.DeepEqual(gotBytes, tt.wantBytes) {
				t.Errorf("StringToBytes() = %v, want %v", gotBytes, tt.wantBytes)
			}
			t.Logf("gotBytes Pointer %p, tt.args.str Pointer %p", &gotBytes, &tt.args.str)
		})
	}
}

func Test_BytesToString(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{
			name: "testBytesToStr",
			args: args{
				bytes: []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"),
			},
			want: "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BytesToString(tt.args.bytes)
			if got != tt.want {
				t.Errorf("BytesToString() = %v, want %v", got, tt.want)
			}
			t.Logf("gotStr Pointer %p, tt.args.bytes Pointer %p", &got, &tt.args.bytes)
		})
	}
}

// go test -v -run=none -bench=. -benchmem=true
// Benchmark_StringToBytes-8       1000000000               0.477 ns/op           0 B/op          0 allocs/op
// Benchmark_getStringToBytes-8    20491503                58.0 ns/op            64 B/op          1 allocs/op
// Benchmark_BytesToString-8       1000000000               0.473 ns/op           0 B/op          0 allocs/op
// Benchmark_getBytesToString-8    22655529                49.1 ns/op            64 B/op          1 allocs/op
var testStr = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
var testBytes = []byte(testStr)

func getBytesToString(bytes []byte) string {
	return string(bytes)
}

func getStringToBytes(str string) []byte {
	return []byte(str)
}

func Benchmark_StringToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		StringToBytes(testStr)
	}
}

func Benchmark_getStringToBytes(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getStringToBytes(testStr)
	}
}

func Benchmark_BytesToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BytesToString(testBytes)
	}
}

func Benchmark_getBytesToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		getBytesToString(testBytes)
	}
}
