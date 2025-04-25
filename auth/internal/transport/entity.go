package transport

var (
	JWTKey string
	JWTCookieName string = "user_jwt"
)

type RegisterRequest struct {
	Login		string	`json:"login"`
	Email		string	`json:"email"`
	Pass		string	`json:"pass"`
	Name		string	`json:"name"`
}

type RegisterResponse struct {
	Token		string	`json:"token"` // TODO: что надо отправлять после регистрации?
}

type LoginRequest struct {
	Credential	string	`json:"credential"` // login ИЛИ email
	Pass		string	`json:"pass"`
}

type LoginResponse struct {
	Token		string	`json:"token"` // TODO: что надо отправлять после входа? То же, что и после регистрации?
}