package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCommandLine(t *testing.T) {
	s := CommandLine("echo", []string{"hello", "world"})
	assert.Equal(t, "echo hello world", s)

	s = CommandLine("kubectl", []string{"get", "pod", "--all-namespaces", "-o", "json"})
	assert.Equal(t, "kubectl get pod --all-namespaces -o json", s)
}

func TestCheckOutput(t *testing.T) {
	err := CheckOutput("hello", "good", "easeprobe hello world")
	assert.Nil(t, err)

	err = CheckOutput("hello", "world", "easeprobe hello world")
	assert.NotNil(t, err)

	err = CheckOutput("hello", "world", "easeprobe hello world")
	assert.NotNil(t, err)

	err = CheckOutput("good", "bad", "easeprobe hello world")
	assert.NotNil(t, err)
}

func TestCheckEmpty(t *testing.T) {
	assert.Equal(t, "a", CheckEmpty("a"))
	assert.Equal(t, "empty", CheckEmpty("    "))
	assert.Equal(t, "empty", CheckEmpty("  \t"))
	assert.Equal(t, "empty", CheckEmpty("\n\r\t"))
	assert.Equal(t, "empty", CheckEmpty("  \n\r\t  "))
}
