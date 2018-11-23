package stringutil_test

import (
	"good/util/stringutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJoinSqlFields(t *testing.T) {
	s := stringutil.JoinSqlFields("", "a", "b")
	assert.Equal(t, "a,b", s)
	s = stringutil.JoinSqlFields("t", "a", "b")
	assert.Equal(t, "t.a,t.b", s)

}
func TestIn(t *testing.T) {
	assert.True(t, stringutil.In(1, []int{1, 2}))
	assert.False(t, stringutil.In(3, []int{1, 2}))
}
func TestFileNameAppend(t *testing.T) {
	assert.Equal(t, "1_1.txt", stringutil.FileNameAppend("1.txt", "_1"))
}
