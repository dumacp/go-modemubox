package modemubox

import (
	"reflect"
	"testing"
)

func Test_extractData(t *testing.T) {
	type args struct {
		data []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "test1",
			args: args{
				data: []string{
					"+CGACT: 1,0",
					"+CGACT: 2,0",
					"+CGACT: 4,1",
					"+CGDCONT: 1,\"IP\",\"nebulaeng-vpn.tigo.com\"",
					"+CGDCONT: 2,\"IP\",\"web.movil.com.co\"",
					"+CGDCONT: 4,\"IP\",\"vpn.tigo.com\"",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractData(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractData() = %q, want %q", got, tt.want)
			}
		})
	}
}

func Test_extractDataWithPrefix(t *testing.T) {
	type args struct {
		data []string
	}
	tests := []struct {
		name string
		args args
		want map[string][]string
	}{
		{
			name: "test1",
			args: args{
				data: []string{
					"+CGACT: 1,0",
					"+CGACT: 2,0",
					"+CGACT: 4,1",
					"+CGDCONT: 1,\"IP\",\"nebulaeng-vpn.tigo.com\"",
					"+CGDCONT: 2,\"IP\",\"web.movil.com.co\"",
					"+CGDCONT: 4,\"IP\",\"vpn.tigo.com\"",
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := extractDataWithPrefix(tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("extractDataWithPrefix() = %v, want %v", got, tt.want)
			}
		})
	}
}
