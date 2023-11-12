package telegram

import (
	"testing"
)

func TestDdCmd(t *testing.T) {
	type args struct {
		txt string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "2 args",
			args: args{txt: "2019-04-05 2023-03-09"},
			want: "1434d (~3.93 года)",
		},
		{
			name: "2 args reverse",
			args: args{txt: "2023-03-09 2019-04-05"},
			want: "1434d (~-3.93 года)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ddCmd(tt.args.txt); got != tt.want {
				t.Errorf("ServiceImpl.ddCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
