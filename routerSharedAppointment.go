/*
@author: 2447899 8689159 3000685
*/
package main

import (
	"DHBW_GO_Projekt/authentifizierung"
	"DHBW_GO_Projekt/terminfindung"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// define Templates
var terminAdminSite = template.Must(template.ParseFiles("./assets/sites/terminfindung/termin-admin.html", "./assets/templates/footer.html", "./assets/templates/header.html"))
var terminSharedCreateLink = template.Must(template.ParseFiles("./assets/sites/terminfindung/termin-create-link.html", "./assets/templates/footer.html", "./assets/templates/header.html"))
var terminSharedCreateLinkPost = template.Must(template.ParseFiles("./assets/sites/terminfindung/termin-showlink.html", "./assets/templates/footer.html", "./assets/templates/header.html"))
var terminSharedCreateDate = template.Must(template.ParseFiles("./assets/sites/terminfindung/termin-create-app.html", "./assets/templates/footer.html", "./assets/templates/header.html"))
var terminSharedShowAllLinks = template.Must(template.ParseFiles("./assets/sites/terminfindung/termin-admin-showAll.html", "./assets/templates/footer.html", "./assets/templates/header.html"))
var errorRoute = template.Must(template.ParseFiles("./assets/sites/error.html", "./assets/templates/footer.html"))
var terminSharedPublic = template.Must(template.ParseFiles("./assets/sites/terminfindung/termin-public.html", "./assets/templates/footer.html", "./assets/templates/header.html"))

// AdminSiteServeHTTP
// handle for /shared --> gives back the overview
func AdminSiteServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//check that user is authenticated and if not redirect to error website and /
	isAllowed, user := checkIfIsAllowed(request)
	if !isAllowed {
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	// get the terminID and selected for get
	termin := request.URL.Query().Get("terminID")
	selectedDay := request.URL.Query().Get("selected")
	// select termin
	terminShared, err := terminfindung.GetTerminFromShared(&user, &termin)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=shared_wrong_terminId&link=" + url.QueryEscape("/user/view")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	//get the selected date if select date is the wrong one it throws an error
	if len(selectedDay) != 0 {
		err := terminfindung.SelectDate(&selectedDay, &termin, &user)
		if err != nil {
			http.Redirect(writer, request, "https://"+request.Host+"/error?type=shared_admin_WrongSelected&link="+url.QueryEscape("/shared?terminID="+termin), http.StatusContinue)
			return
		}
		// get teh right termin
		terminShared, _ = terminfindung.GetTerminFromShared(&user, &termin)
	}
	terminForHTML := terminShared.ConvertAdminToHTML()
	// execute the admin path always if not (on error) it redirectes to error
	err = terminAdminSite.Execute(writer, terminForHTML)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		fmt.Println(err)
		return
	}
}

// CreateLinkServeHTTP
// for createing a link
func CreateLinkServeHTTP(writer http.ResponseWriter, request *http.Request) {
	//check if user is allowed
	isAllowed, user := checkIfIsAllowed(request)
	if !isAllowed {
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	// get the terminID
	termin := request.URL.Query().Get("terminID")
	//if the data needs to be parsed then it automatically changes the structs
	if request.Method == http.MethodPost {
		// parse forms
		err := request.ParseForm()
		if err != nil {
			request.Method = "GET"
			urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/shared?terminID="+termin)
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		// get the name of the user that should be created
		name := request.Form.Get("name")
		linkForPerson, err := terminfindung.CreatePerson(&name, &termin, &user)
		//if metho is wrong it automatically says no
		if err != nil {
			request.Method = "GET"
			urls := "https://" + request.Host + "/error?type=shared_coudntCreatePerson&link=" + url.QueryEscape("/shared?terminID="+termin)
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		// make custom struckt for user
		type links struct {
			LinkForUser string
			LinkBack    string
		}
		linksFortemp := links{
			LinkBack:    "/shared?terminID=" + termin,
			LinkForUser: "https://" + request.Host + "/shared/public?" + linkForPerson,
		}
		err = terminSharedCreateLinkPost.Execute(writer, linksFortemp)
		if err != nil {
			request.Method = "GET"
			urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/shared?terminID="+termin)
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		return
	}
	err := terminSharedCreateLink.Execute(writer, termin)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/shared?terminID="+termin)
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
}

// ServeHTTPSharedAppCreateDate
// Creates appointemntn with whole dates
func ServeHTTPSharedAppCreateDate(writer http.ResponseWriter, request *http.Request) {
	isAllowed, user := checkIfIsAllowed(request)
	if !isAllowed {
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	termin := request.URL.Query().Get("terminID")
	if request.Method == http.MethodPost {
		//if is submited the form it pases the request
		err := request.ParseForm()
		if err != nil {
			request.Method = "GET"
			urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/shared?terminID="+termin)
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		// get start date and enddate and parse it to time format
		startDate := request.Form.Get("startdate")
		endDate := request.Form.Get("enddate")
		startDateFormated, err := time.Parse("2006-01-02", startDate)
		enddateFormated, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			request.Method = "GET"
			urls := "https://" + request.Host + "/error?type=wrong_date_format&link=" + url.QueryEscape("/shared?terminID="+termin)
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		//create a new proposed date and redirect to the main website
		err = terminfindung.CreateNewProposedDate(startDateFormated, enddateFormated, &user, &termin, false)
		if err != nil {
			request.Method = "GET"
			urls := "https://" + request.Host + "/error?type=dateIsAfter&link=" + url.QueryEscape("/shared/create/app?terminID="+termin)
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		//on post redirect to main website
		request.Method = "GET"
		http.Redirect(writer, request, "https://"+request.Host+"/shared?terminID="+termin, http.StatusContinue)
		return
	}
	_, err := terminfindung.GetTerminFromShared(&user, &termin)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/shared?terminID="+termin)
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	err = terminSharedCreateDate.Execute(writer, termin)
	if err != nil {
		log.Println("Coudn't Execute Template in Create Date")
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}

}

// ShowAllLinksServeHttp
// shows all links of user and termin if it is authenticate
func ShowAllLinksServeHttp(writer http.ResponseWriter, request *http.Request) {
	// check whether user is allowed to access site
	isAllowed, userAdmin := checkIfIsAllowed(request)
	if !isAllowed {
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	// get the terminID and check whether one can acces it
	termin := request.URL.Query().Get("terminID")
	links, err := terminfindung.GetAllLinks(&userAdmin, &termin)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	// get the links and then add it to local struct
	for key, user := range links {
		links[key].Url = "https://" + request.Host + "/shared/public?terminID=" + url.QueryEscape(termin) + "&name=" + user.Name + "&user=" + userAdmin + "&apiKey=" + user.Url

	}
	//setup struct for html template
	type shared struct {
		Users     []terminfindung.UserTermin
		Routeback string
	}
	//add object for templates
	forTemplate := shared{
		Users:     links,
		Routeback: termin,
	}
	// execute
	err = terminSharedShowAllLinks.Execute(writer, forTemplate)
	if err != nil {
		log.Println("Coudn't Execute Template in show all links")
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	return
}

// PublicSharedWebsite
// public website for every user without authentication to access --> needs api key
func PublicSharedWebsite(writer http.ResponseWriter, request *http.Request) {
	var apikey string
	//on post it edits it
	if request.Method == http.MethodPost {
		// check form and if not give back internal error
		err := request.ParseForm()
		if err != nil {
			log.Println("coudn't parse form")
			urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/")
			http.Redirect(writer, request, urls, http.StatusContinue)
			return
		}
		//get the apikey date key and voted --> in form is beeing added by js --> so i dont need to load it in get
		apikey = url.QueryEscape(request.Form.Get("apiKey"))
		dateKey := request.Form.Get("dateKey")
		voted := request.Form.Get("voted")
		votedBool := false
		// https gives back on if checkbox is checked --> sp it checks whether it is the sam
		if strings.Compare(voted, "on") == 0 {
			votedBool = true
		}
		termin, user, err := terminfindung.GetTerminViaApiKey(&apikey)
		if err != nil {
			return
		}
		err = terminfindung.VoteForDay(&termin.Info.ID, &termin.User, &user, &dateKey, votedBool)
		if err != nil {
			return
		}
	} else {
		// if not post it is set in the apikey --> is a long byte hash
		apikey = url.QueryEscape(request.URL.Query().Get("apiKey"))
	}
	// gets the termin via api key
	termin, user, err := terminfindung.GetTerminViaApiKey(&apikey)
	if err != nil {
		urls := "https://" + request.Host + "/error?type=wrongAuthentication&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	htmlInput := termin.ConvertUserSiteToRightHTML(&user, &apikey)
	err = terminSharedPublic.Execute(writer, htmlInput)
	if err != nil {
		log.Println("coudn't execute form in all links")
		urls := "https://" + request.Host + "/error?type=internal&link=" + url.QueryEscape("/")
		http.Redirect(writer, request, urls, http.StatusContinue)
		return
	}
	return
}

func checkIfIsAllowed(request *http.Request) (isAllowed bool, username string) {
	cookie, err := request.Cookie("SessionID-Kalender")
	//if cookie is not existing it returns back to the host
	if err != nil {
		isAllowed = false
		return
	}
	isAllowed, username = authentifizierung.CheckCookie(&cookie.Value)
	return
}
