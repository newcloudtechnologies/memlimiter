/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2022.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package tracker

import (
	"testing"

	"github.com/newcloudtechnologies/memlimiter/utils/config/duration"
)

func TestConfigBackendFile_Prepare(t *testing.T) {
	type fields struct {
		Path string
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "empty path",
			fields:  fields{Path: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ConfigBackendFile{
				Path: tt.fields.Path,
			}

			err := c.Prepare()
			if (err != nil) != tt.wantErr {
				t.Errorf("Prepare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestConfig_Prepare(t *testing.T) {
	type fields struct {
		BackendFile   *ConfigBackendFile
		BackendMemory *ConfigBackendMemory
		Period        duration.Duration
	}

	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "empty backends",
			fields:  fields{},
			wantErr: true,
		},
		{
			name: "non-empty backends",
			fields: fields{
				BackendFile:   new(ConfigBackendFile),
				BackendMemory: new(ConfigBackendMemory),
			},
			wantErr: true,
		},
		{
			name: "invalid duration",
			fields: fields{
				BackendFile: new(ConfigBackendFile),
				Period:      duration.Duration{Duration: 0},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Config{
				BackendFile:   tt.fields.BackendFile,
				BackendMemory: tt.fields.BackendMemory,
				Period:        tt.fields.Period,
			}

			err := c.Prepare()
			if (err != nil) != tt.wantErr {
				t.Errorf("Prepare() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
