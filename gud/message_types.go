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
	Title string `json:"title"`
	Content string `json:"content"`
}

type status uint

const (
	IOpen status = 0
	IInprogress status = 1
	IDone status = 2
	IClose status = 3
)

type Issue struct {
	Id int `json:"id"`
	Title string `json:"title"`
	Author string `json:"author"`
	Content string `json:"content"`
	Status status `json:"status"`
}

type GetIssuesResponse struct {
	Issues []Issue `json:"issues"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MultiErrorResponse struct {
	Errors []string `json:"errors"`
}

const PasswordLenMin = 8

