package modemubox

import "testing"

func Test_getCidAndIP(t *testing.T) {
	type args struct {
		s []string
	}
	tests := []struct {
		name    string
		args    args
		want    int
		want1   string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				s: []string{"AT+CGCONTRDP",
					"",
					"+CGCONTRDP: 4,5,\"nebulaeng-vpn.tigo.com.MNC103.MCC732.GPRS\",\"10.2.183.132.255.255.255.255\",\"10.2.183.132\",\"0.0.0.0\",\"0.0.0.0\",\"0.0.0.0\",\"0.0.0.0\",0",
					"",
					"OK"},
			},
			want:    4,
			want1:   "10.2.183.132",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := getCidAndIP(tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("getCidAndIP() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getCidAndIP() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("getCidAndIP() got1 = %v, want %v", got1, tt.want1)
			}
			t.Logf("cid: %d, ip: %q\n", got, got1)
		})
	}
}
