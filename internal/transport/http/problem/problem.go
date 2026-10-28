package problem

type Document struct {
	Status int    `json:"status"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

func New(status int, title, detail string) Document {
	return Document{
		Status: status,
		Title:  title,
		Detail: detail,
	}
}
