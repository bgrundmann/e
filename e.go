package main
import "github.com/nsf/termbox-go"

func draw(b *Buf) {
	const coldef = termbox.ColorDefault
	termbox.Clear(coldef, coldef)
	//_w, _h := termbox.Size()
	termbox.Flush()
} 

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	} 
	defer termbox.Close()
	termbox.SetInputMode(termbox.InputEsc)
	var b Buf
	b.Init()
	b.Insert(0, []byte("World"))
	b.Insert(0, []byte("Hello"))
	b.Insert(5, []byte(" "))

mainloop:
	for {
		draw(&b)
		switch ev := termbox.PollEvent(); ev.Type {
			case termbox.EventKey:
				switch ev.Key {
					case termbox.KeyEsc:
						break mainloop

					default:
				}
			case termbox.EventError:
				panic(ev.Err)
		} 
	} 
}
