package global

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetWritableDir(t *testing.T) {
	filename := ""
	dir := MakeDirectory(filename)
	assert.Equal(t, GetWorkDir(), dir)

	filename = "./test.txt"
	dir = MakeDirectory(filename)
	exp, _ := filepath.Abs(filename)
	assert.Equal(t, exp, dir)

	filename = "./none/existed/test.txt"
	exp, _ = filepath.Abs(filename)
	dir = MakeDirectory(filename)
	os.RemoveAll("./none")
	assert.Equal(t, exp, dir)

	filename = "~/none/existed/test.txt"
	exp = filepath.Join(os.Getenv("HOME"), "none/existed/test.txt")
	dir = MakeDirectory(filename)
	os.RemoveAll(os.Getenv("HOME") + "/none")
	assert.Equal(t, exp, dir)
}
