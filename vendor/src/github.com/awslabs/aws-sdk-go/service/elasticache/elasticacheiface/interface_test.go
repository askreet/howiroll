// THIS FILE IS AUTOMATICALLY GENERATED. DO NOT EDIT.

package elasticacheiface_test

import (
	"testing"

	"github.com/awslabs/aws-sdk-go/service/elasticache"
	"github.com/awslabs/aws-sdk-go/service/elasticache/elasticacheiface"
	"github.com/stretchr/testify/assert"
)

func TestInterface(t *testing.T) {
	assert.Implements(t, (*elasticacheiface.ElastiCacheAPI)(nil), elasticache.New(nil))
}