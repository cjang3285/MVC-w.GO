package greetings

import (
	"testing"
	"regexp"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestHelloName(t *testing.T){
	name := "Jang chanwook"
	want := regexp.MustCompile(`\b`+name+`\b`) // not error but my fault 1 :`\b` <- '\b'
	msg, err := Hello("Jang chanwook")
	if !want.MatchString(msg) || err != nil {
		t.Errorf(`Hello("Jang chanwook") = %q, %v, want match for %#q, nil`, msg, err, want)
	}
}

//TestHelloEmpty calls greetings.Hello with empty string,
//checking for an error.
func TestHelloEmpty(t *testing.T){
	msg, err := Hello("")
	if msg != "" || err == nil {
		t.Errorf(`Hello("") = %q, %v, want "", error`, msg, err) //same fault : ` ` <- ""
	}
}
