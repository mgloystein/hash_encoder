package server

type messageResult struct {
	Message string `json:"message"`
}

type stats struct {
	Count              int64   `json:"total"`
	AverageProcessTime float64 `json:"average"`
}