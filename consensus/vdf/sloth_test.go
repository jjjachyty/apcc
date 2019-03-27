package vdf

import "testing"

func TestFixed_delay(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name string
		args args
	}{
		{"test", args{[]string{"9853393445385562019", "83271827105964338786165203944215195008542731998507103719200925244436566851770", "500000"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Fixed_delay(tt.args.args)
		})
	}
}
