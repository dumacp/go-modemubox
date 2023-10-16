package main

import (
	"bufio"
	"reflect"
	"testing"
)

type MockPort struct {
	Data bufio.Reader
}

func (mp *MockPort) Read(p []byte) (n int, err error) {

	data := make([]byte, len(p))
	copy(data, p)
	// Implement your mock Read function here
	return len(data), nil
}

func (mp *MockPort) Write(p []byte) (n int, err error) {
	// Implement your mock Write function here
	return len(p), nil
}

func Test_getCtxActive(t *testing.T) {
	type args struct {
		res []string
	}
	tests := []struct {
		name string
		args args
		want map[int]int
	}{
		{
			name: "test1",
			args: args{
				// res: []string{"+CGACT: 1,0", "+CGACT: 2,0", "+CGACT: 4,1"},
				res: []string{"1,0", "2,0", "4,1"},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parsecontext(tt.args.res); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getCtxActive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseapn(t *testing.T) {
	type args struct {
		res []string
	}
	tests := []struct {
		name string
		args args
		want map[int]int
	}{
		{
			name: "test1",
			args: args{
				// res: []string{"+CGACT: 1,0", "+CGACT: 2,0", "+CGACT: 4,1"},
				res: []string{"1,\"IP\",\"nebulaeng-vpn.tigo.com\",\"\",0,0,0,0,0,0",
					"2,\"IP\",\"web.colombiamovil.com.co\",\"\",0,0,0,0,0,0",
					"3,\"IP\",\"\",\"\",0,0,0,0,0,0",
					"4,\"IP\",\"nebulaeng-vpn.tigo.com.MNC103.MCC732.GPRS\",\"10.2.183.132\",0,0,0,0,0,0"},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseapn(tt.args.res); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseapn() = %v, want %v", got, tt.want)
			}
		})
	}
}
