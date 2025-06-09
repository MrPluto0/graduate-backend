package algorithm

type State struct {
	T        uint     `json:"t"`         // 系统时间
	UserData []uint   `json:"user_data"` // 该时间每个用户产生的需处理任务数据
	CommData [][]uint `json:"comm_data"` // 该时间分配给每个通信设备需要处理的用户数据
	ProcData [][]uint `json:"proc_data"` // 该时间每个通信设备处理的用户数据
	Queue    []uint   `json:"queue"`     // 每个通信设备的队列长度
}
