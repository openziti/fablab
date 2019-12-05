package console

type Message struct {
	Source  string    `json:"source"`
	Routers []string  `json:"routers"`
	Links   []*Link   `json:"links"`
	Metrics []*Metric `json:"metrics"`
}

type Link struct {
	Id  string `json:"id"`
	Src string `json:"src"`
	Dst string `json:"dst"`
}

type Metric struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}
