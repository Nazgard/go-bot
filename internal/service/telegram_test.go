package service

import "testing"

func TestServiceImpl_ddCmd(t *testing.T) {
	type args struct {
		txt string
	}
	tests := []struct {
		name string
		s    *ServiceImpl
		args args
		want string
	}{
		{
			name: "2 args",
			s:    nil,
			args: args{txt: "2019-04-05 2023-03-09"},
			want: "1434d (~3.93 года)",
		},
		{
			name: "2 args reverse",
			s:    nil,
			args: args{txt: "2023-03-09 2019-04-05"},
			want: "1434d (~-3.93 года)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.ddCmd(tt.args.txt); got != tt.want {
				t.Errorf("ServiceImpl.ddCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
