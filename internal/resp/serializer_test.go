package resp

import (
	"errors"
	"testing"
)

func TestArray(t *testing.T) {
	type args struct {
		elements []interface{}
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "all strings",
			args: args{
				elements: []interface{}{"DEL", "a", "key"},
			},
			want: "*3\r\n$3\r\nDEL\r\n$1\r\na\r\n$3\r\nkey\r\n",
		},

		{
			name: "all ints",
			args: args{
				elements: []interface{}{1, 2},
			},
			want: "*2\r\n:1\r\n:2\r\n",
		},

		{
			name: "mixed",
			args: args{
				elements: []interface{}{"GET", 2},
			},
			want: "*2\r\n$3\r\nGET\r\n:2\r\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Array(tt.args.elements); got != tt.want {
				t.Errorf("Array() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError(t *testing.T) {
	ErrDummy := errors.New("dummy Error")

	gotErr := Error(ErrDummy)

	wantErr := "-Dummy Error\r\n"
	if gotErr != wantErr {
		t.Errorf("Error(): %v , want: %v", gotErr, wantErr)
	}
}
