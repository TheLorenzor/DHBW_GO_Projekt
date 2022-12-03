package main

import (
	"DHBW_GO_Projekt/authentifizierung"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var mainRoute = template.Must(template.ParseFiles("./assets/sites/index.html", "./assets/templates/footer.html"))
var createUserRouter = template.Must(template.ParseFiles("./assets/sites/user-create.html", "./assets/templates/footer.html", "./assets/templates/header.html"))
var changeUserRoute = template.Must(template.ParseFiles("./assets/sites/user-change.html", "./assets/templates/footer.html", "./assets/templates/header.html"))
var userOverview = template.Must(template.ParseFiles("./assets/sites/user.html", "./assets/templates/footer.html", "./assets/templates/header.html"))

func (h RootHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	if request.Method == "POST" {
		err := request.ParseForm()
		if err != nil {
			return
		}
		username := request.Form.Get("user")
		password := request.Form.Get("password")
		// cookie authentifizieren checken
		isUser, cookieText := authentifizierung.AuthenticateUser(&username, &password)
		if isUser == true {
			// wenn user authentifiziert ist dann wird cookie erstellt und
			cookie := &http.Cookie{
				Name:     "SessionID-Kalender",
				Value:    cookieText,
				Path:     "/",
				MaxAge:   3600,
				Secure:   true,
				SameSite: http.SameSiteLaxMode,
			}
			http.SetCookie(writer, cookie)
			//redirect to new site
			http.Redirect(writer, request, "https://"+request.Host+"/user/view", http.StatusFound)
			return
		} else {
			// wenn nicht authentifiziert ist wird weiter geleitet oder bei problemen gibt es ein 500 status
			if len(cookieText) == 0 {
				writer.WriteHeader(500)
			} else {
				request.Method = "GET"
				urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
				http.Redirect(writer, request, urls, http.StatusContinue)
				return
			}
		}
	}
	err := mainRoute.Execute(writer, nil)
	if err != nil {
		log.Fatal("Coudnt export Parsefiles")
	}

}

func (createUser CreatUserHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// get cookie
	isAllowed, _ := checkIfIsAllowed(request)
	if !isAllowed {
		http.Redirect(writer, request, "https://"+request.Host, http.StatusContinue)
		return
	}
	//if it is post it should process the data
	if request.Method == "POST" {
		// if the parseform isnt correct it should return
		err := request.ParseForm()
		if err != nil {
			http.Redirect(writer, request, "https://"+request.Host, http.StatusContinue)
		}
		//get the user from the form request
		user := request.Form.Get("newUsername")
		password := request.Form.Get("newPassword")
		err = authentifizierung.CreateUser(&user, &password)
		if err != nil {
			http.Redirect(writer, request, "https://"+request.Host, http.StatusContinue)

		}
		//if successfull on post it should return back to the user
		http.Redirect(writer, request, "https://"+request.Host, http.StatusContinue)

	}

	err := createUserRouter.Execute(writer, nil)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/user")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}

}

func (changeUser ChangeUserHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	isAllowed, user := checkIfIsAllowed(request)
	if !isAllowed {
		http.Redirect(writer, request, "https://"+request.Host, http.StatusContinue)
		return
	}
	if request.Method == "POST" {
		//if post request it actually parses the form and trys to change the password and create a new cookie
		err := request.ParseForm()
		if err != nil {
			return
		}
		//change the user to new user
		password := request.Form.Get("oldPassword")
		newPassword := request.Form.Get("newPassword")
		cookies, err := authentifizierung.ChangeUser(&user, &password, &newPassword)
		if err != nil {
			urls := "https://" + request.Host + "/error?type=wrongPassword&link=" + url.QueryEscape("/user")
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		// set cookie so it automatically updates and it doesnt throw one back to the login site
		cookie := &http.Cookie{
			Name:     "SessionID-Kalender",
			Value:    cookies,
			Path:     "/",
			MaxAge:   3600,
			Secure:   true,
			SameSite: http.SameSiteLaxMode,
		}
		//set new cookie and redirect
		http.SetCookie(writer, cookie)
		http.Redirect(writer, request, "https://"+request.Host+"/user", http.StatusContinue)
		return

	}
	//execute own template from userchange and put in footer and header
	err := changeUserRoute.Execute(writer, nil)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
}

func (user UserHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	isAllowed, username := checkIfIsAllowed(request)
	if !isAllowed {
		request.Method = "GET"
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	err := userOverview.Execute(writer, username)
	if err != nil {
		request.Method = "GET"
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
}

// handleTableView
// Hier werden all http-Request anfragen geregelt, die im Kontext der TableView anfallen
func (v *ViewmanagerHandler) handleTableView(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		switch {
		case r.RequestURI == "/user/view/table?suche=minusMonat":
			v.vm.TvJumpMonthBack()
		case r.RequestURI == "/user/view/table?suche=plusMonat":
			v.vm.TvJumpMonthFor()
		case strings.Contains(r.RequestURI, "/user/view/table?monat="):
			monatStr := r.FormValue("monat")
			monat, _ := strconv.Atoi(monatStr)
			v.vm.TvSelectMonth(time.Month(monat))
		case r.RequestURI == "/user/view/table?jahr=Zurueck":
			v.vm.TvJumpYearForOrBack(-1)
		case r.RequestURI == "/user/view/table?jahr=Vor":
			v.vm.TvJumpYearForOrBack(1)
		case r.RequestURI == "/user/view/table?datum=heute":
			v.vm.TvJumpToToday()
		case strings.Contains(r.RequestURI, "/user/view/table/editor"):
			terminToEdit := v.vm.GetTerminInfos(r)
			er := v.viewmanagerTpl.ExecuteTemplate(w, "editor.html", terminToEdit)
			if er != nil {
				log.Fatalln(er)
			}
			return
		}
	}

	if r.Method == "POST" {
		switch {
		case r.RequestURI == "/user/view/table?terminErstellen":
			v.vm.CreateTermin(r, v.vm.Username)
		case strings.Contains(r.RequestURI, "/user/view/table/editor"):
			v.vm.EditTermin(r, v.vm.Username)
		}
	}

	er := v.viewmanagerTpl.ExecuteTemplate(w, "tbl.html", v.vm.Tv)
	if er != nil {
		log.Fatalln(er)
	}
}

// ListHandler
// Hier werden all http-Request-Anfragen geregelt, die im Kontext der Listenansicht anfallen
func (v *ViewmanagerHandler) handleListView(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		switch {
		case strings.Contains(r.RequestURI, "/user/view/list?selDate="):
			dateStr := r.FormValue("selDate")
			v.vm.LvSelectDate(dateStr)
		case strings.Contains(r.RequestURI, "/user/view/list?Eintraege="):
			amountStr := r.FormValue("Eintraege")
			amount, _ := strconv.Atoi(amountStr)
			v.vm.LvSelectEntriesPerPage(amount)
		case r.RequestURI == "/user/view/list?Seite=Vor":
			v.vm.LvJumpPageForward()
		case r.RequestURI == "/user/view/list?Seite=Zurueck":
			v.vm.LvJumpPageBack()
		case strings.Contains(r.RequestURI, "/user/view/list/editor"):
			terminToEdit := v.vm.GetTerminInfos(r)
			er := v.viewmanagerTpl.ExecuteTemplate(w, "editor.html", terminToEdit)
			if er != nil {
				log.Fatalln(er)
			}
			return
		}
	}

	if r.Method == "POST" {
		switch {
		case r.RequestURI == "/user/view/list?terminErstellen":
			v.vm.CreateTermin(r, v.vm.Username)
		case strings.Contains(r.RequestURI, "/user/view/list/editor"):
			v.vm.EditTermin(r, v.vm.Username)
		}
	}
	er := v.viewmanagerTpl.ExecuteTemplate(w, "liste.html", v.vm.Lv)
	if er != nil {
		log.Fatalln(er)
	}
}
func (v *ViewmanagerHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//TODO implement me
	panic("implement me")
}

func (l LogoutHandler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	cookie, err := request.Cookie("SessionID-Kalender")
	if err != nil {
		http.Redirect(writer, request, "https://"+request.Host, http.StatusContinue)
		return

	}

	cookie.Value = ""
	cookie.Expires = time.Unix(0, 0)
	http.SetCookie(writer, cookie)
	http.Redirect(writer, request, "https://"+request.Host, http.StatusContinue)
}
