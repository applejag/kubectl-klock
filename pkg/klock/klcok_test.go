package klock

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOptions_NormalizedLabelColumns(t *testing.T) {
	labelColumnsTests := []struct {
		labelColumns           []string
		normalizedLabelColumns []string
	}{
		{
			labelColumns:           []string{"app"},
			normalizedLabelColumns: []string{"app"},
		},
		{
			labelColumns:           []string{"app,version"},
			normalizedLabelColumns: []string{"app", "version"},
		},
		{
			labelColumns:           []string{"app", "version"},
			normalizedLabelColumns: []string{"app", "version"},
		},
		{
			labelColumns:           []string{"app", "version,role"},
			normalizedLabelColumns: []string{"app", "version", "role"},
		},
		{
			labelColumns:           []string{" app , version ", " role   "},
			normalizedLabelColumns: []string{"app", "version", "role"},
		},
	}

	for _, labelColumnsTest := range labelColumnsTests {
		o := Options{LabelColumns: labelColumnsTest.labelColumns}
		assert.Exactly(t, labelColumnsTest.normalizedLabelColumns, o.NormalizedLabelColumns())
	}
}
