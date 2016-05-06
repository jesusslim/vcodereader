package main

import (
	"fmt"
	"github.com/jesusslim/vcodereader"
)

func main() {
	file_name := "test.png"
	vr := vcodereader.NewVcodeReaderDefault(file_name)
	r, err := vr.Read()
	if err != nil {
		fmt.Println(err.Error)
	} else {
		fmt.Println(r)
	}
}
