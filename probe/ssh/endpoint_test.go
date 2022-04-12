package ssh

import "testing"

func check(t *testing.T, fname string, err error, result, expected string) {
	if err != nil {
		t.Errorf("%s error: %s", fname, err)
	}
	if result != expected {
		t.Errorf("%s result: %s, expected: %s", fname, result, expected)
	}
}

func TestParseHost(t *testing.T) {
	e := Endpoint{}

	const fname = "ParseHost"
	e.Host = "example.com:22"
	check(t, fname, e.ParseHost(), e.Host, "example.com:22")

	e.Host = "192.168.1.1"
	check(t, fname, e.ParseHost(), e.Host, "192.168.1.1:22")

	e.Host = "example.com:2222"
	check(t, fname, e.ParseHost(), e.Host, "example.com:2222")

	e.Host = "user@example.com"
	check(t, fname, e.ParseHost(), e.Host, "example.com:22")
	check(t, fname, nil, e.User, "user")
}
