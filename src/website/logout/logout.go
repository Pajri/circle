package logout

import (
	"net/http"

	utils "../../utils"
)

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	s, err := utils.GetSession(r, utils.SESSION_AUTH) //init session
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "logout : an error occured when getting session.")
		return
	}

	//check is logged in
	isAuth := utils.IsLoggedInSession(s)
	if !isAuth {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	err = utils.Logout(s, w, r)
	if err != nil {
		utils.InternalServerErrorHandler(w, r, err, "logout : an error occured when logging out.")
		return
	}
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}
