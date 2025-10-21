package define

// TransferPath 传输路径（包含路径、速率、功率等预计算信息）
type TransferPath struct {
	Path   []uint    `json:"path"`   // 设备ID路径
	Speeds []float64 `json:"speeds"` // 每段的传输速率
	Powers []float64 `json:"powers"` // 每段的传输功率
}

// Copy 深拷贝
func (tp *TransferPath) Copy() *TransferPath {
	if tp == nil {
		return nil
	}
	return &TransferPath{
		Path:   append([]uint(nil), tp.Path...),
		Speeds: append([]float64(nil), tp.Speeds...),
		Powers: append([]float64(nil), tp.Powers...),
	}
}
