package osutil_test

import (
	"good/util/osutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMkDir(t *testing.T) {
	d := "_TestMkDir/1"
	err := osutil.MkDir(d)
	assert.Nil(t, err)
	os.RemoveAll("_TestMkDir")
}
