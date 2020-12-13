package resp

import (
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"reflect"
	"testing"
)

func TestDeserialize(t *testing.T) {
	type args struct {
		serializedCMD string
	}
	tests := []struct {
		name        string
		args        args
		wantCMD     common.CommandID
		wantCMDArgs common.CommandArguments
		wantErr     bool
	}{
		{
			name:    "SET",
			args:    args{serializedCMD: " SET foo 1"},
			wantCMD: common.SET,
			wantCMDArgs: common.SETArguments{
				Key:   "foo",
				Value: "1",
			},
			wantErr: false,
		},
		{
			name:    "SET with enclosing quotes ",
			args:    args{serializedCMD: `SET   full-name "John Doe" `},
			wantCMD: common.SET,
			wantCMDArgs: common.SETArguments{
				Key:   "full-name",
				Value: "John Doe",
			},
			wantErr: false,
		},
		{
			name:    "GET",
			args:    args{serializedCMD: " GET     foo"},
			wantCMD: common.GET,
			wantCMDArgs: common.GETArguments{
				Key: "foo",
			},
			wantErr: false,
		},
		{
			name:    "DEL",
			args:    args{serializedCMD: "DEL   foo"},
			wantCMD: common.DEL,
			wantCMDArgs: common.DELArguments{
				Keys: []string{"foo"},
			},
			wantErr: false,
		},

		{
			name:    "DEL with multiple keys",
			args:    args{serializedCMD: "DEL key1 key2 key3"},
			wantCMD: common.DEL,
			wantCMDArgs: common.DELArguments{
				Keys: []string{"key1", "key2", "key3"},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCMD, gotCMDArgs, err := Deserialize(tt.args.serializedCMD)
			if (err != nil) != tt.wantErr {
				t.Errorf("Deserialize() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCMD != tt.wantCMD {
				t.Errorf("Deserialize() got = %v, want %v", gotCMD, tt.wantCMD)
			}
			if !reflect.DeepEqual(gotCMDArgs, tt.wantCMDArgs) {
				t.Errorf("Deserialize() got1 = %v, want %v", gotCMDArgs, tt.wantCMDArgs)
			}
		})
	}
}
