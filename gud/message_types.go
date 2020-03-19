package gud

type SignUpRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type CreateProjectRequest struct {
	Name string `json:"name"`
}

type InviteMemberRequest struct {
	Name string `json:"name"`
}

type CreateIssueRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

type CreatePrRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type CreateIssueResponse struct {
	Id int `json:"id"`
}

type Issue struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Content string `json:"content"`
	Status  string `json:"status"`
}

type PullRequest struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Author  string `json:"author"`
	Content string `json:"content"`
	From    string `json:"from"`
	To      string `json:"to"`
}

type Job struct {
	Id      int    `json:"id"`
	Version string `json:"version"`
	Status  string `json:"status"`
	Logs    string `json:"logs,omitempty"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MultiErrorResponse struct {
	Errors []string `json:"errors"`
}

const PasswordLenMin = 8
