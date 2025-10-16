package handlers

type AddToUGroupBody struct {
	Username string   `json:"username" binding:"required"`
	Groups   []string `json:"groups"   binding:"required,min=1,dive,required"`
}

type AddToRGroupBody struct {
	Resource string   `json:"username" binding:"required"`
	Groups   []string `json:"groups"   binding:"required,min=1,dive,required"`
}
