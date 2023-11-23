package redis_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/osmosis-labs/osmosis/v20/ingest/sqs/domain/mvc"
)

var poolRepo mvc.PoolsRepository

func TestRedis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redis Suite")
}
