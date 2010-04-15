package stomp

import (
	"bytes"
	"testing"
)

func TestFrameStruct(t *testing.T) {

	command := "SUBSCRIBE"
	body := "Some body."
	h := header{
		"destination": "/queue/test",
		"other-header": "foobar"}

	fString := command + "\n"
	for name, value := range h {
		fString += name + ": " + value + "\n"
	}
	fString += "\n" + body

	f, err := frameFromString(fString)
	if err != nil {
		t.Errorf("error %s", err)
	}

	if f.command != command {
		t.Errorf("expected command '%s', got '%s'", command, f.command)
	}
	if f.header["destination"] != "/queue/test" {
		t.Errorf("couldn't find header destination '/queue/test'")
	}
	if f.header["other-header"] != "foobar" {
		t.Errorf("couldn't find header other-header 'foobar'")
	}
	if f.body != body {
		t.Errorf("expected body '%s', got '%s'", body, f.body)
	}

	// convert to string and back again
	var b = new(bytes.Buffer)
	f.writeTo(b)
	f2, err := frameFromString(b.String())
	if err != nil {
		t.Errorf("error %s", err)
	}

	if f2.command != command {
		t.Errorf("expected command '%s', got '%s'", command, f2.command)
	}
	if f2.header["destination"] != "/queue/test" {
		t.Errorf("couldn't find header destination '/queue/test'")
	}
	if f2.header["other-header"] != "foobar" {
		t.Errorf("couldn't find header other-header 'foobar'")
	}
	if f2.body != body {
		t.Errorf("expected body '%s', got '%s'", body, f2.body)
	}

}
