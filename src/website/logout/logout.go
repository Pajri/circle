package logout

import (
	"net/http"

	utils "../../utils"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	err := utils.Logout(w, r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "logout : an error occured when logging out.")
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
