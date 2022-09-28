package spec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseMetricsFile(t *testing.T) {
	var sample = `entities: []`
	spec, err := ParseMetricsFile([]byte(sample))
	assert.Nil(t, err)
	assert.NotNil(t, spec.Entities)
}
