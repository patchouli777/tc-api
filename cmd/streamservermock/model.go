package main

type livestream struct {
	channel string
	viewers int
}

type StreamData struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Vhost     string `json:"vhost"`
	App       string `json:"app"`
	TcURL     string `json:"tcUrl"`
	URL       string `json:"url"`
	LiveMs    int64  `json:"live_ms"`
	Clients   int    `json:"clients"`
	Frames    int    `json:"frames"`
	SendBytes int    `json:"send_bytes"`
	RecvBytes int    `json:"recv_bytes"`
	Kbps      struct {
		Recv30S int `json:"recv_30s"`
		Send30S int `json:"send_30s"`
	} `json:"kbps"`
	Publish struct {
		Active bool   `json:"active"`
		Cid    string `json:"cid"`
	} `json:"publish"`
	Video struct {
		Codec   string `json:"codec"`
		Profile string `json:"profile"`
		Level   string `json:"level"`
		Width   int    `json:"width"`
		Height  int    `json:"height"`
	} `json:"video"`
	Audio struct {
		Codec      string `json:"codec"`
		SampleRate int    `json:"sample_rate"`
		Channel    int    `json:"channel"`
		Profile    string `json:"profile"`
	} `json:"audio"`
}
