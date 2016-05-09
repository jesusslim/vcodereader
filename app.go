package vcodereader

import (
	"fmt"
	"github.com/otiai10/gosseract"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
)

const (
	TYPE_PNG  = 1
	TYPE_JPEG = 2
)

type VcodeReader struct {
	lineInfos    []*LineInfo //干扰线组
	width        int         //图片宽度
	height       int         //图片高度
	file_name    string      //文件名
	file_type    int         //文件类型
	prefix       string      //生成的新图片前缀
	max_linked   int         //需删除的干扰线的最小长度
	max_jump     int         //干扰线断开跨度
	check_round2 bool        //是否检查每个节点第二圈
	xy_arr_out   []*Xy       //外圈偏移量
	m            image.Image
	use_client   bool
	need_rev     bool //是否灰度反转
}

//第一圈和第二圈偏移量
var xy_arr_round1 []*Xy
var xy_arr_round2 []*Xy

func init() {
	for i := -1; i < 2; i++ {
		for j := -1; j < 2; j++ {
			if i == 0 && j == 0 {
				continue
			}
			xy_arr_round1 = append(xy_arr_round1, NewXy(i, j))
		}
	}
	for i := -2; i < 3; i++ {
		for j := -2; j < 3; j++ {
			if i >= -1 && i <= 1 && j >= -1 && j <= 1 {
				continue
			}
			xy_arr_round2 = append(xy_arr_round2, NewXy(i, j))
		}
	}
}

func NewVcodeReader(file_name string, check_round2 bool, max_linked int, max_jump int, save_clear_file_name_prefix string, use_client bool, need_rev bool) *VcodeReader {
	vr := &VcodeReader{
		lineInfos:    []*LineInfo{},
		width:        0,
		height:       0,
		file_name:    file_name,
		prefix:       save_clear_file_name_prefix,
		max_linked:   max_linked,
		max_jump:     max_jump,
		check_round2: check_round2,
		xy_arr_out:   []*Xy{},
		use_client:   use_client,
		need_rev:     need_rev,
	}
	suffix := file_name[strings.LastIndex(file_name, ".")+1:]
	switch suffix {
	case "jpg", "jpeg":
		vr.file_type = TYPE_JPEG
		panic("This File is not support.")
		break
	case "png":
		vr.file_type = TYPE_PNG
		break
	default:
		panic("This File is not support.")
	}
	var border int
	if check_round2 {
		border = 2
	} else {
		border = 1
	}
	for i := -max_jump; i < max_jump+1; i++ {
		for j := -max_jump; j < max_jump+1; j++ {
			if i >= -border && i <= border && j >= -border && j <= border {
				continue
			}
			vr.xy_arr_out = append(vr.xy_arr_out, NewXy(i, j))
		}
	}
	return vr
}

func NewVcodeReaderDefault(file_name string) *VcodeReader {
	return NewVcodeReader(file_name, true, 3, 6, "clear_", false, false)
}

func (this *VcodeReader) SetNeedRev(need_rev bool) {
	this.need_rev = need_rev
}

func (this *VcodeReader) Read() (string, error) {
	source := this.file_name
	save_name := this.prefix + this.file_name
	f0, ferr := os.Create(save_name)
	if ferr != nil {
		return "", ferr
	}
	f1, serr := os.OpenFile(source, os.O_RDWR|os.O_APPEND|os.O_CREATE, os.ModeType)
	if serr != nil {
		return "", serr
	}
	defer f1.Close()
	var m1 image.Image
	var err error
	if this.file_type == TYPE_PNG {
		m1, err = png.Decode(f1)
	} else {
		m1, err = jpeg.Decode(f1)
	}
	if err != nil {
		return "", err
	}
	this.m = m1
	bound := this.m.Bounds()
	this.width = bound.Dx()
	this.height = bound.Dy()
	//find other lines
	//找出干扰线
	for x := 0; x < this.width; x++ {
		xy := NewXy(x, 0)
		lineInfoNow := NewLineInfo(xy)
		this.lineInfos = append(this.lineInfos, lineInfoNow)
		this.tracePoints(1, lineInfoNow, nil, 0, 0, 0)

		xy2 := NewXy(x, this.height-1)
		lineInfoNow2 := NewLineInfo(xy2)
		this.lineInfos = append(this.lineInfos, lineInfoNow2)
		this.tracePoints(2, lineInfoNow2, nil, 0, 0, 0)
	}
	fmt.Println("40%")
	for y := 0; y < this.height; y++ {
		xy := NewXy(0, y)
		lineInfoNow := NewLineInfo(xy)
		this.lineInfos = append(this.lineInfos, lineInfoNow)
		this.tracePoints(-1, lineInfoNow, nil, 0, 0, 0)

		xy2 := NewXy(this.width-1, y)
		lineInfoNow2 := NewLineInfo(xy2)
		this.lineInfos = append(this.lineInfos, lineInfoNow2)
		this.tracePoints(-2, lineInfoNow2, nil, 0, 0, 0)
	}
	fmt.Println("80%")
	rgba := image.NewRGBA(image.Rect(0, 0, this.width, this.height))
	//灰度化
	//灰度反转 *
	sum_blue := uint32(0)
	sum_count := uint32(0)
	for y := 0; y < this.height; y++ {
		for x := 0; x < this.width; x++ {
			r, g, b, a := this.m.At(x, y).RGBA()
			r, g, b, a = r>>8, g>>8, b>>8, a>>8
			if this.need_rev {
				r, g, b = 255-r, 255-g, 255-b
			}
			rgba.Set(x, y, color.NRGBA{uint8(r * 30 / 100), uint8(g * 59 / 100), uint8(b * 11 / 100), uint8(a)})
			//rgba.Set(x, y, color.NRGBA{uint8(255), uint8(255), uint8(255), uint8(a)})
			//fmt.Println("OLD:", r, ",", g, ",", b, ",", a)
			_, _, b2, _ := rgba.At(x, y).RGBA()
			sum_blue += b2
			sum_count++
			//r2, g2, b2, a2 = r2>>8, g2>>8, b2>>8, a2>>8
			//fmt.Println("NEW:", r2, ",", g2, ",", b2, ",", a2)
		}
	}
	//二值化
	avg_blue := sum_blue / sum_count
	for y := 0; y < this.height; y++ {
		for x := 0; x < this.width; x++ {
			_, _, b, a := rgba.At(x, y).RGBA()
			a = a >> 8
			if b > avg_blue {
				rgba.Set(x, y, color.NRGBA{uint8(255), uint8(255), uint8(255), uint8(a)})
			} else {
				rgba.Set(x, y, color.NRGBA{uint8(0), uint8(0), uint8(0), uint8(a)})
			}
		}
	}
	//去除干扰点 TODO round1
	var xy_arr [4]Xy
	xy_arr[0] = Xy{-1, 0}
	xy_arr[1] = Xy{1, 0}
	xy_arr[2] = Xy{0, -1}
	xy_arr[3] = Xy{0, 1}
	for y := 0; y < this.height; y++ {
		for x := 0; x < this.width; x++ {
			_, _, _, a := rgba.At(x, y).RGBA()
			if a > 0 {
				need_clear := true
				for _, tmp := range xy_arr {
					_, _, _, a_t := rgba.At(x+tmp.x, y+tmp.y).RGBA()
					if a_t > 0 {
						need_clear = false
						break
					}
				}
				if need_clear {
					rgba.Set(x, y, color.NRGBA{uint8(255), uint8(255), uint8(255), uint8(255)})
				}
			}
		}
	}
	//去除干扰线
	for _, info := range this.lineInfos {
		if info.LenPoints() >= this.max_linked {
			for _, p := range info.points {
				rgba.Set(p.x, p.y, color.NRGBA{uint8(255), uint8(255), uint8(255), uint8(255)})
			}
		}
	}
	if this.file_type == TYPE_PNG {
		err = png.Encode(f0, rgba)
	} else {
		err = jpeg.Encode(f0, rgba, &jpeg.Options{5})
	}
	if err != nil {
		return "", err
	}
	if this.use_client {
		client, _ := gosseract.NewClient()
		r, e := client.Src(save_name).Out()
		if e != nil {
			return "", e
		} else {
			return strings.TrimSpace(r), e
		}
	} else {
		out := gosseract.Must(gosseract.Params{
			Src:       save_name,
			Languages: "eng+heb",
		})
		return strings.TrimSpace(out), nil
	}
}

func (this *VcodeReader) tracePoints(direction int, lineInfo *LineInfo, nowXy *Xy, r, g, b uint32) bool {
	m := this.m
	width := this.width
	height := this.height
	times := 0
	if nowXy == nil {
		nowXy = lineInfo.from
		r, g, b, _ = m.At(nowXy.x, nowXy.y).RGBA()
	}
	// if r == 65535 && g == 65535 && b == 65535 {
	// 	return true
	// }
	lineInfo.points = append(lineInfo.points, nowXy)
	lineInfoHere := lineInfo.Copy()
	for _, xy := range xy_arr_round1 {
		if direction == -1 {
			//横向 向右
			if xy.x < 0 {
				continue
			}
		} else if direction == 1 {
			//纵向 向下
			if xy.y < 0 {
				continue
			}
		} else if direction == -2 {
			//横向 向左
			if xy.x > 0 {
				continue
			}
		} else if direction == 2 {
			//纵向 向上
			if xy.y > 0 {
				continue
			}
		}
		w := nowXy.x + xy.x
		h := nowXy.y + xy.y
		if w >= 0 && w < width && h >= 0 && h < height {
			next_xy := NewXy(w, h)
			r2, g2, b2, _ := m.At(next_xy.x, next_xy.y).RGBA()
			if r2 == r && g2 == g && b2 == b {
				//防止倒回
				if lineInfoHere.PointExist(next_xy.x, next_xy.y) {
					continue
				}
				if times == 0 {
					return this.tracePoints(direction, lineInfo, next_xy, r, g, b)
				} else {
					lineInfoNew := lineInfoHere.Copy()
					this.lineInfos = append(this.lineInfos, lineInfoNew)
					return this.tracePoints(direction, lineInfoNew, nowXy, r, g, b)
				}
				times++
			}
		}
	}
	if times == 0 && this.check_round2 {
		for _, xy := range xy_arr_round2 {
			if direction == -1 {
				//横向 向右
				if xy.x < 0 {
					continue
				}
			} else if direction == 1 {
				//纵向 向下
				if xy.y < 0 {
					continue
				}
			} else if direction == -2 {
				//横向 向左
				if xy.x > 0 {
					continue
				}
			} else if direction == 2 {
				//纵向 向上
				if xy.y > 0 {
					continue
				}
			}
			w := nowXy.x + xy.x
			h := nowXy.y + xy.y
			if w >= 0 && w < width && h >= 0 && h < height {
				next_xy := NewXy(w, h)
				r2, g2, b2, _ := m.At(next_xy.x, next_xy.y).RGBA()
				if r2 == r && g2 == g && b2 == b {
					//防止倒回
					if lineInfoHere.PointExist(next_xy.x, next_xy.y) {
						continue
					}
					if times == 0 {
						return this.tracePoints(direction, lineInfo, next_xy, r, g, b)
					} else {
						lineInfoNew := lineInfoHere.Copy()
						this.lineInfos = append(this.lineInfos, lineInfoNew)
						return this.tracePoints(direction, lineInfoNew, nowXy, r, g, b)
					}
					times++
				}
			}
		}
	}
	if times == 0 {
		//完全断开 找n格以内同颜色 重新开
		for _, xy := range this.xy_arr_out {
			if direction == -1 {
				//横向 向右
				if xy.x < 0 {
					continue
				}
			} else if direction == 1 {
				//纵向 向下
				if xy.y < 0 {
					continue
				}
			} else if direction == -2 {
				//横向 向左
				if xy.x > 0 {
					continue
				}
			} else if direction == 2 {
				//纵向 向上
				if xy.y > 0 {
					continue
				}
			}
			w := nowXy.x + xy.x
			h := nowXy.y + xy.y
			if w >= 0 && w < width && h >= 0 && h < height {
				next_xy := NewXy(w, h)
				r2, g2, b2, _ := m.At(next_xy.x, next_xy.y).RGBA()
				if r2 == r && g2 == g && b2 == b {
					//防止倒回
					if lineInfoHere.PointExist(next_xy.x, next_xy.y) {
						continue
					}
					lineInfoNew := NewLineInfo(next_xy)
					this.lineInfos = append(this.lineInfos, lineInfoNew)
					return this.tracePoints(direction, lineInfoNew, nil, r, g, b)
				}
			}
		}
	}
	return true
}
