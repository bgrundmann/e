package buf

import "io"
import "bufio"
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

func ExampleBufReader() {
	var b Buf
	b.Init()
	b.Insert(0, []byte("Hello"))
	r := bufio.NewReaderSize(b.NewReader(0), 128)
	s, err := r.ReadString('\n')
	if err != io.EOF {
		fmt.Printf("expected EOF", err)
	}
	fmt.Printf("%s\n", s)
	// Output: Hello
}

func TestBufReverseReader(t *testing.T) {
	var b Buf
	b.Init()
	b.Insert(0, []byte("Hello"))
	r := b.NewReader(5)
	r.Reverse()
	check := func(c rune) {
		if ch, n, err := r.ReadRune(); !(ch == c && n == 1 && err == nil) {
			t.Errorf("Expected %c got: %c", c, ch)	
		} 
	} 
	check('o')
	check('l')
	check('l')
	check('e')
	check('H')
	if ch, n, err := r.ReadRune(); err != io.EOF {
		t.Errorf("Expected EOF got: %c - %i - %v", ch, n, err)
	} 
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

func TestLine(t *testing.T) {
	var b Buf
	b.Init()
	b.Insert(0, []byte("Hello\nWorld\n\nThis is a test\n"))
	test := func(n, off int) {
		got := b.Line(n)
		if got != off {
			t.Errorf("Line %v expected %v got: %v", n, off, got)
		}
	}
	test(1, 0)
	test(2, 6)
	test(3, 12)
	test(4, 13)
}

func TestLines(t *testing.T) {
	var b Buf
	b.Init()
	if b.Lines() != 1 {
		t.Errorf("empty buffer should have 1 line")
	}
	b.Insert(0, []byte("Hello\n\nFoo"))
	if n := b.Lines(); n != 3 {
		t.Errorf("expected 3 lines got %v", n)
	}
}
