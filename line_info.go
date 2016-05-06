package vcodereader

//干扰线
type LineInfo struct {
	from   *Xy
	points []*Xy
}

func (l *LineInfo) GetRoot() *Xy {
	return l.from
}

func (l *LineInfo) GetPoints() []*Xy {
	return l.points
}

func (l *LineInfo) LenPoints() int {
	return len(l.points)
}

func (l *LineInfo) PointExist(x, y int) bool {
	for _, xy := range l.points {
		if xy.x == x && xy.y == y {
			return true
		}
	}
	return false
}

func (l *LineInfo) Copy() *LineInfo {
	copy := NewLineInfo(l.from)
	for _, p := range l.points {
		copy.points = append(copy.points, p)
	}
	return copy
}

func NewLineInfo(xy *Xy) *LineInfo {
	return &LineInfo{
		from:   xy,
		points: []*Xy{},
	}
}
