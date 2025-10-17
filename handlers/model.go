package handlers

type UserCreateBody struct {
	ID       string `binding:"omitempty,uuid4" json:"id"`
	Username string `binding:"required" json:"username"`
	Password string `binding:"required" json:"password"`
}

type UserUpdateBody struct {
	ID          string `binding:"omitempty,uuid4" json:"id"`
	NewUsername string `binding:"omitempty" json:"newUsername"`
	NewPassword string `binding:"omitempty" json:"newPassword"`
}

type UserBody struct {
	ID string `binding:"required,uuid4" json:"id"`
}

type AddToUGroupBody struct {
	Username string   `binding:"required"                     json:"username"`
	Groups   []string `binding:"required,min=1,dive,required" json:"groups"`
}

type AddToRGroupBody struct {
	Resource string   `binding:"required"                     json:"resource"`
	Groups   []string `binding:"required,min=1,dive,required" json:"groups"`
}

type CreatePolicyBody struct {
	UserGroup     string `binding:"required" json:"userGroup"`
	ResourceGroup string `binding:"required" json:"resourceGroup"`
	Action        string `binding:"required" json:"action"`
}

type ResponseError struct {
	Error string `json:"error"`
}
