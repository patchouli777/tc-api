package streamserver

type ListResponse struct {
	Code    int          `json:"code"`
	Server  string       `json:"server"`
	Service string       `json:"service"`
	Pid     string       `json:"pid"`
	Streams []StreamData `json:"streams"`
}

type GetResponse struct {
	Code    int        `json:"code"`
	Server  string     `json:"server"`
	Service string     `json:"service"`
	Pid     string     `json:"pid"`
	Stream  StreamData `json:"stream"`
}

type PostRequest struct {
	Channel string `json:"channel"`
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

type StreamEventPayload struct {
	ServerID  string `json:"server_id"`
	Action    string `json:"action"` // e.g. "on_publish", "on_unpublish"
	ClientID  string `json:"client_id"`
	IP        string `json:"ip"`
	Vhost     string `json:"vhost"`
	App       string `json:"app"`
	TcURL     string `json:"tcUrl"`
	Stream    string `json:"stream"` // channel
	Param     string `json:"param"`
	StreamURL string `json:"stream_url"`
	StreamID  string `json:"stream_id"`
}

type SubscribeRequest struct {
	CallbackURL string
}

type SubscribeResponse struct {
	Success bool
}

type UnsubscribeRequest struct {
	CallbackURL string
}
