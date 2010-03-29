package stomp

import (
	"testing"
)

func TestFrameStruct(t *testing.T) {

	command := "SUBSCRIBE"
	body := "Some body."
	headers := map[string]string{
		"destination": "/queue/test",
		"other-header": "foobar"}

	fString := command + "\n"
	for name, value := range headers {
		fString += name + ": " + value + "\n"
	}
	fString += "\n" + body

	f := frameFromString(fString)

	if f.command != command {
		t.Errorf("expected command '%s', got '%s'", command, f.command)
	}
	if f.headers["destination"] != "/queue/test" {
		t.Errorf("couldn't find header destination '/queue/test'")
	}
	if f.headers["other-header"] != "foobar" {
		t.Errorf("couldn't find header other-header 'foobar'")
	}
	if f.body != body {
		t.Errorf("expected body '%s', got '%s'", body, f.body)
	}

	// convert to string and back again
	fString2 := f.string()
	f2 := frameFromString(fString2)

	if f2.command != command {
		t.Errorf("expected command '%s', got '%s'", command, f2.command)
	}
	if f2.headers["destination"] != "/queue/test" {
		t.Errorf("couldn't find header destination '/queue/test'")
	}
	if f2.headers["other-header"] != "foobar" {
		t.Errorf("couldn't find header other-header 'foobar'")
	}
	if f2.body != body {
		t.Errorf("expected body '%s', got '%s'", body, f2.body)
	}

}
