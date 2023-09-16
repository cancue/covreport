package reporter_test

import (
	"testing"

	"github.com/cancue/covreport/reporter"
	"github.com/stretchr/testify/assert"
)

func TestParseCutlines(t *testing.T) {
	t.Run("should return error when cannot parse safe cutlines", func(t *testing.T) {
		var err error
		_, err = reporter.ParseCutlines("")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("not-a-number,5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3,not-a-number")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3, 5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3 ,5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines(" 3,5")
		assert.ErrorContains(t, err, "invalid syntax")

		_, err = reporter.ParseCutlines("3,5 ")
		assert.ErrorContains(t, err, "invalid syntax")
	})

	t.Run("should return last number as warning cut", func(t *testing.T) {
		cutlines, err := reporter.ParseCutlines("3")
		assert.NoError(t, err)
		assert.Equal(t, 3.0, cutlines.Safe)
		assert.Equal(t, 3.0, cutlines.Warning)

		cutlines, err = reporter.ParseCutlines("3,5")
		assert.NoError(t, err)
		assert.Equal(t, 3.0, cutlines.Safe)
		assert.Equal(t, 5.0, cutlines.Warning)

		cutlines, err = reporter.ParseCutlines("3,5,7")
		assert.NoError(t, err)
		assert.Equal(t, 3.0, cutlines.Safe)
		assert.Equal(t, 7.0, cutlines.Warning)
	})
}
