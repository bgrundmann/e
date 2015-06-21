package main

import "fmt"
import "testing"

func ExampleBufInsert() {
	var b Buf
	b.Init()
	b.Insert(0, []byte("World"))
	b.Insert(0, []byte("Hello"))
	b.Insert(5, []byte(" "))
	fmt.Printf("%s\n", &b)
	// Output: Hello World
}

func ExampleBufDelete() {
	var b Buf
	b.Init()
	b.Insert(0, []byte("Hello"))
	b.Delete(0, b.Len())
	fmt.Printf("%s\n", &b)
	// Output:
}

func TestDeleteEnd(t *testing.T) {
	var b Buf
	b.Init()
	b.Insert(0, []byte("Hello"))
	b.Delete(3, b.Len())
	s := b.String()
	if s != "Hel" {
		t.Errorf("expected: \"Hel\" got: %q", s)
	}
}

func TestDeleteStart(t *testing.T) {
	var b Buf
	b.Init()
	b.Insert(0, []byte("Hello"))
	b.Delete(0, 2)
	s := b.String()
	if s != "llo" {
		t.Errorf("expected: \"llo\" got: %q", s)
	}
}

func TestDeleteStartEnd(t *testing.T) {
	var b Buf
	b.Init()
	b.Insert(0, []byte("Hello"))
	b.Delete(2, 3)
	s := b.String()
	if s != "Helo" {
		t.Errorf("expected: \"Helo\" got: %q", s)
	}
}
