package vcodereader

//坐标点
type Xy struct {
	x int
	y int
}

func (this *Xy) GetX() int {
	return this.x
}

func (this *Xy) GetY() int {
	return this.y
}

func NewXy(x, y int) *Xy {
	return &Xy{x, y}
}
