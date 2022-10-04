/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package utils

import (
	"testing"
)

func TestClampFloat64(t *testing.T) {
	type args struct {
		value float64
		min   float64
		max   float64
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "less",
			args: args{
				value: -1,
				min:   0,
				max:   100,
			},
			want: 0,
		},
		{
			name: "middle",
			args: args{
				value: 50,
				min:   0,
				max:   100,
			},
			want: 50,
		},
		{
			name: "greater",
			args: args{
				value: 101,
				min:   0,
				max:   100,
			},
			want: 100,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ClampFloat64(tt.args.value, tt.args.min, tt.args.max); got != tt.want {
				t.Errorf("ClampFloat64() = %v, want %v", got, tt.want)
			}
		})
	}
}
