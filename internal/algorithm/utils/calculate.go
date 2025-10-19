package utils

import (
	"go-backend/internal/algorithm/constant"
	"math"
)

// 两点的距离
func Distance(x1, y1, x2, y2 float64) float64 {
	return math.Hypot(x2-x1, y2-y1)
}

// TransferSpeed 计算传输速率（基于自由空间路径损失模型和香农公式）
// p_t: 发射功率 (W)
// d: 传输距离 (m)
// 返回: 传输速率 (bits/s)
func TransferSpeed(p_t float64, d float64) float64 {
	// 波长 λ = c / f
	lamb := 3e8 / constant.Wireless

	// 自由空间路径损失模型
	// 接收功率 P_r = (P_t * λ²) / (4πd)²
	denominator := 4 * math.Pi * d
	p_r := (p_t * lamb * lamb) / (denominator * denominator)

	// 香农公式: C = B * log2(1 + SNR)
	// SNR = P_r / Noise
	v := constant.Bdw * math.Log2(1+(p_r/constant.Noise))

	return v
}
