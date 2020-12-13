package resp

import (
	"bufio"
	"github.com/rilopez/redis-wire-protocol/internal/common"
	"net/textproto"
	"reflect"
	"strings"
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
			args:    args{serializedCMD: "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\n100\r\n"},
			wantCMD: common.SET,
			wantCMDArgs: common.SETArguments{
				Key:   "foo",
				Value: "100",
			},
			wantErr: false,
		},
		{
			name:    "SET with enclosing quotes ",
			args:    args{serializedCMD: "*3\r\n$3\r\nSET\r\n$9\r\nfull-name\r\n$10\r\n\"John Doe\"\r\n"},
			wantCMD: common.SET,
			wantCMDArgs: common.SETArguments{
				Key:   "full-name",
				Value: "\"John Doe\"",
			},
			wantErr: false,
		},
		//{
		//	name:    "GET",
		//	args:    args{serializedCMD: " GET     foo"},
		//	wantCMD: common.GET,
		//	wantCMDArgs: common.GETArguments{
		//		Key: "foo",
		//	},
		//	wantErr: false,
		//},
		//{
		//	name:    "DEL",
		//	args:    args{serializedCMD: "DEL   foo"},
		//	wantCMD: common.DEL,
		//	wantCMDArgs: common.DELArguments{
		//		Keys: []string{"foo"},
		//	},
		//	wantErr: false,
		//},
		//
		//{
		//	name:    "DEL with multiple keys",
		//	args:    args{serializedCMD: "*\r\nDEL key1 key2 key3"},
		//	wantCMD: common.DEL,
		//	wantCMDArgs: common.DELArguments{
		//		Keys: []string{"key1", "key2", "key3"},
		//	},
		//	wantErr: false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := bufio.NewReader(strings.NewReader(tt.args.serializedCMD))
			tp := textproto.NewReader(r)
			gotCMD, gotCMDArgs, err := DeserializeCMD(tp)
			if (err != nil) != tt.wantErr {
				t.Errorf("DeserializeCMD() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotCMD != tt.wantCMD {
				t.Errorf("DeserializeCMD() gotCMD = %v, wantCMD %v", gotCMD, tt.wantCMD)
			}
			if !reflect.DeepEqual(gotCMDArgs, tt.wantCMDArgs) {
				t.Errorf("DeserializeCMD() gotCMDArgs = %v, wantCMDArgs %v", gotCMDArgs, tt.wantCMDArgs)
			}
		})
	}
}
