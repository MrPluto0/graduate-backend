package constant

// 系统基本参数
const (
	// 设备高度，单位：米
	H = 30
	// 系统时隙，单位：秒
	Slot = 0.05
	// 通信半径，单位：米
	Radius = 400
	// 资源利用率
	Eta = 0.9
)

// 控制参数
const (
	// 迭代次数
	Iters = 20
	// 控制参数V
	V = 100
	// 收缩参数
	Shrink = 10000
	// 偏置参数
	Bias = 0.1
	// 权重参数α
	Alpha = 0.3
	// 权重参数β
	Beta = 0.3
	// 权重参数γ
	Gamma = 0.4
)

// 计算相关参数
const (
	// CPU频率，单位：Hz（1GHz）
	C = 1e9
	// 计算1比特所需转数，单位：转/比特
	Rho = 1000
)

// 传输相关参数
const (
	// 用户功率，单位：W（4G LTE 0.1~0.2W）
	P_u = 0.2
	// 基站功率，单位：W
	P_b = 2
	// 通信带宽，单位：Hz（10MHz）
	Bdw = 5e7
	// FR1无线频率，单位：Hz
	Wireless = 3.5e9
	// 噪声功率，单位：W
	Noise = 1e-9
)

// 能耗相关参数
const (
	// 计算能耗参数
	Kappa = 10e-28
	// 保持状态能耗，单位：W/s
	E_hold = 220
)
