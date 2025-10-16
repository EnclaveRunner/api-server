package handlers

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
