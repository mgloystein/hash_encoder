package common

type MessageResult struct {
	Message string `json:"message"`
}

type Stats struct {
	Count              int64   `json:"total"`
	AverageProcessTime float64 `json:"average"`
}
