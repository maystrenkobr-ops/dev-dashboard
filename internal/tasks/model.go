package tasks

type Task struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	Priority  string `json:"priority"`
	Deadline  string `json:"deadline"`
	CreatedAt string `json:"created_at"`
}
