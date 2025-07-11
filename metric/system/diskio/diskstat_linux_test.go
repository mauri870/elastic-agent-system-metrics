// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//go:build !integration && linux
// +build !integration,linux

package diskio

import (
	"math"
	"testing"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/stretchr/testify/assert"

	"github.com/elastic/elastic-agent-system-metrics/metric"
)

func Test_GetCLKTCK(t *testing.T) {
	//usually the tick is 100
	assert.Equal(t, uint32(100), GetCLKTCK())
}

func Test32BitRollover(t *testing.T) {
	var maxUint32 uint64 = math.MaxUint32 // 4_294_967_295

	var prev = maxUint32 - 100_000

	// A rolled-over value
	var current32 uint64 = 1000
	// Theoretical un-rolled over value
	var current64 = (maxUint32 + current32)

	var correct = current64 - prev
	assert.Equal(t, returnOrFix32BitRollover(current32, prev), returnOrFix32BitRollover(current64, prev))
	assert.Equal(t, correct, returnOrFix32BitRollover(current32, prev))

	assert.Equal(t, uint64(0), returnOrFix32BitRollover(current32, current32))
}

func TestDiskIOStat_CalIOStatistics(t *testing.T) {
	counter := disk.IOCountersStat{
		ReadCount:  13,
		WriteCount: 17,
		ReadTime:   19,
		WriteTime:  23,
		Name:       "iostat",
	}

	stat := &IOStat{
		lastDiskIOCounters: map[string]disk.IOCountersStat{
			"iostat": {
				ReadCount:  3,
				WriteCount: 5,
				ReadTime:   7,
				WriteTime:  11,
				Name:       "iostat",
			},
		},
		lastCPU: cpu.TimesStat{Idle: 1},
		curCPU:  cpu.TimesStat{Idle: 0.01},
	}

	expected := IOMetric{
		AvgAwaitTime:      metric.Round(24.0 / 22.0),
		AvgReadAwaitTime:  1.2,
		AvgWriteAwaitTime: 1,
	}

	got, err := stat.CalcIOStatistics(counter)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, expected.AvgAwaitTime, got.AvgAwaitTime)
	assert.Equal(t, expected.AvgReadAwaitTime, got.AvgReadAwaitTime)
	assert.Equal(t, expected.AvgWriteAwaitTime, got.AvgWriteAwaitTime)
}
