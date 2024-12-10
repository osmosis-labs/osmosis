package types

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sqspassthroughdomain "github.com/osmosis-labs/osmosis/v28/ingest/types/passthroughdomain"
	api "github.com/osmosis-labs/sqs/pkg/api/v1beta1/pools"
)

func TestPoolWrapper_Incentive(t *testing.T) {
	tests := []struct {
		name     string
		aprData  sqspassthroughdomain.PoolAPRDataStatusWrap
		expected api.IncentiveType
	}{
		{
			name: "Superfluid Incentive",
			aprData: sqspassthroughdomain.PoolAPRDataStatusWrap{
				PoolAPR: sqspassthroughdomain.PoolAPR{
					SuperfluidAPR: sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
					OsmosisAPR:    sqspassthroughdomain.PoolDataRange{Lower: 0, Upper: 0},
				},
			},
			expected: api.IncentiveType_SUPERFLUID,
		},
		{
			name: "Osmosis Incentive",
			aprData: sqspassthroughdomain.PoolAPRDataStatusWrap{
				PoolAPR: sqspassthroughdomain.PoolAPR{
					SuperfluidAPR: sqspassthroughdomain.PoolDataRange{Lower: 0, Upper: 0},
					OsmosisAPR:    sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
					BoostAPR:      sqspassthroughdomain.PoolDataRange{Lower: 0, Upper: 0},
				},
			},
			expected: api.IncentiveType_OSMOSIS,
		},
		{
			name: "Boost Incentive",
			aprData: sqspassthroughdomain.PoolAPRDataStatusWrap{
				PoolAPR: sqspassthroughdomain.PoolAPR{
					BoostAPR: sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
				},
			},
			expected: api.IncentiveType_BOOST,
		},
		{
			name: "No Incentive",
			aprData: sqspassthroughdomain.PoolAPRDataStatusWrap{
				PoolAPR: sqspassthroughdomain.PoolAPR{
					SuperfluidAPR: sqspassthroughdomain.PoolDataRange{Lower: 0, Upper: 0},
					OsmosisAPR:    sqspassthroughdomain.PoolDataRange{Lower: 0, Upper: 0},
					BoostAPR:      sqspassthroughdomain.PoolDataRange{Lower: 0, Upper: 0},
				},
			},
			expected: api.IncentiveType_NONE,
		},
		{
			name: "Multiple Incentives - Superfluid Priority",
			aprData: sqspassthroughdomain.PoolAPRDataStatusWrap{
				PoolAPR: sqspassthroughdomain.PoolAPR{
					SuperfluidAPR: sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
					OsmosisAPR:    sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
					BoostAPR:      sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
				},
			},
			expected: api.IncentiveType_SUPERFLUID,
		},
		{
			name: "Multiple Incentives - Osmosis Priority",
			aprData: sqspassthroughdomain.PoolAPRDataStatusWrap{
				PoolAPR: sqspassthroughdomain.PoolAPR{
					SuperfluidAPR: sqspassthroughdomain.PoolDataRange{Lower: 0, Upper: 0},
					OsmosisAPR:    sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
					BoostAPR:      sqspassthroughdomain.PoolDataRange{Lower: 1, Upper: 2},
				},
			},
			expected: api.IncentiveType_OSMOSIS,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pool := &PoolWrapper{
				APRData: tt.aprData,
			}
			assert.Equal(t, tt.expected, pool.Incentive())
		})
	}
}
