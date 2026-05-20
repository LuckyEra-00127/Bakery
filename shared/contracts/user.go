package contracts

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at,omitempty"`
}

type RegisterUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Role     string `json:"role,omitempty"`
}

type LoginUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type ValidateTokenRequest struct {
	Token string `json:"token"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type GetUserByIDRequest struct {
	ID string `json:"id"`
}

type GetUserByEmailRequest struct {
	Email string `json:"email"`
}

type UpdateUserProfileRequest struct {
	ID       string `json:"id"`
	FullName string `json:"full_name"`
}

type ChangePasswordRequest struct {
	ID          string `json:"id"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type CreateStoreProfileRequest struct {
	UserID    string `json:"user_id"`
	StoreName string `json:"store_name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

type UpdateStoreProfileRequest struct {
	UserID    string `json:"user_id"`
	StoreName string `json:"store_name"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

type ListClientsRequest struct{}

type AssignRoleRequest struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
}

type DeleteUserRequest struct {
	ID string `json:"id"`
}

type UserResponse struct {
	User    User   `json:"user"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
}

type LoginResponse struct {
	User         User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type ValidateTokenResponse struct {
	Valid bool   `json:"valid"`
	User  User   `json:"user,omitempty"`
	Error string `json:"error,omitempty"`
}

type ListClientsResponse struct {
	Clients []User `json:"clients"`
}
