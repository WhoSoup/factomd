package state

import (
	"reflect"
	"testing"
)

func mmU(in ...[]uint32) [][]uint32 {
	return in
}
func mU(in ...uint32) []uint32 {
	return in
}
func Test_partition(t *testing.T) {
	type args struct {
		n []uint32
	}
	tests := []struct {
		name string
		args args
		want [][]uint32
	}{
		{"empty input", args{nil}, [][]uint32{}},
		{"single digit", args{mU(1)}, mmU(mU(1, 1))},
		{"scattered 2", args{mU(1, 3)}, mmU(mU(1, 1), mU(3, 3))},
		{"1-2", args{mU(1, 2)}, mmU(mU(1, 2))},
		{"1-3", args{mU(1, 2, 3)}, mmU(mU(1, 3))},
		{"1-3 & 5", args{mU(1, 2, 3, 5)}, mmU(mU(1, 3), mU(5, 5))},
		{"1-3 & 5-7", args{mU(1, 2, 3, 5, 6, 7)}, mmU(mU(1, 3), mU(5, 7))},
		{"1-10", args{mU(1, 2, 3, 4, 5, 6, 7, 8, 9, 10)}, mmU(mU(1, 10))},
		{"1-3 & 5-7 & 9-12", args{mU(1, 2, 3, 5, 6, 7, 9, 10, 11, 12)}, mmU(mU(1, 3), mU(5, 7), mU(9, 12))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := partition(tt.args.n); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("partition() = %v, want %v", got, tt.want)
			}
		})
	}
}
