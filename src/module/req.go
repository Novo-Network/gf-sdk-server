package module

type PutObject struct {
	Data        string `json:"data"`
	ContentType string `json:"content_type"`
	Visibility  int32  `json:"visibility"`
	Sync        bool   `json:"sync"`
}

func (o *PutObject) Check() {
	if o.Visibility == 0 {
		o.Visibility = 3
	}

	if o.ContentType == "" {
		o.ContentType = "application/octet-stream"
	}
}
