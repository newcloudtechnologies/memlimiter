/*
 * Copyright (c) New Cloud Technologies, Ltd. 2013-2026.
 * Author: Vitaly Isaev <vitaly.isaev@myoffice.team>
 * License: https://github.com/newcloudtechnologies/memlimiter/blob/master/LICENSE
 */

package nextgc

import (
	"testing"

	configbytes "github.com/newcloudtechnologies/memlimiter/utils/config/bytes"
	"github.com/stretchr/testify/require"
)

func TestComputeGoAllocLimit(t *testing.T) {
	tests := []struct {
		name         string
		rssLimit     uint64
		cgoAllocs    uint64
		expected     uint64
		expectedOkay bool
	}{
		{
			name:         "regular budget",
			rssLimit:     1000,
			cgoAllocs:    100,
			expected:     900,
			expectedOkay: true,
		},
		{
			name:         "budget exhausted exactly",
			rssLimit:     1000,
			cgoAllocs:    1000,
			expected:     1,
			expectedOkay: false,
		},
		{
			name:         "budget exhausted overflow risk",
			rssLimit:     1000,
			cgoAllocs:    1200,
			expected:     1,
			expectedOkay: false,
		},
		{
			name:         "empty rss limit",
			rssLimit:     0,
			cgoAllocs:    0,
			expected:     0,
			expectedOkay: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &controllerImpl{
				cfg: &ControllerConfig{
					RSSLimit: configbytes.Bytes{Value: tt.rssLimit},
				},
			}

			actual, ok := c.computeGoAllocLimit(tt.cgoAllocs)

			require.Equal(t, tt.expected, actual)
			require.Equal(t, tt.expectedOkay, ok)
		})
	}
}
