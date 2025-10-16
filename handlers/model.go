package handlers

type AddToGroupBody struct {
	Username string   `json:"username" binding:"required"`
	Groups   []string `json:"groups"   binding:"required,min=1,dive,required"`
}
