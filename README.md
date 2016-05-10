# vcodereader
A golang tool to read verify code.

# How to use
1. install [tesseract-ocr](https://code.google.com/p/tesseract-ocr/)
2. install [go](http://golang.org/doc/install)
3. install [gosseract](https://godoc.org/github.com/otiai10/gosseract)
    - `go get github.com/otiai10/gosseract`
4. install [vcodereader](https://godoc.org/github.com/jesusslim/vcodereader)
    - `go get github.com/jesusslim/vcodereader`
5. run the example:vcodereader/example/example.go

# 如何使用
使用步骤参考上述 
依赖于tesseract-ocr、otiai10/gosseract以及go运行环境
上述依赖安装完毕后 使用go get命令获取此项目代码
运行example文件夹下例子

# Example code:

	package main

	import (
		"fmt"
		"github.com/jesusslim/slimgo/utils"
		"github.com/jesusslim/vcodereader"
		"time"
	)

	func main() {
		fmt.Println(utils.TimeFormat(time.Now(), "H:i:s"))
		//file_name := "test.png"
		//vr := vcodereader.NewVcodeReaderDefault(file_name)
		vr := vcodereader.NewVcodeReaderDefaultFromUrl("yoururl")
		//vr.SetNeedRev(true)
		r, err := vr.Read()
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(r)
		}
		fmt.Println(utils.TimeFormat(time.Now(), "H:i:s"))
	}
