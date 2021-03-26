package main

import (
	"html/template"
	_ "net/http/pprof"
)

var tmpl httptemplate

type httptemplate struct {
	shopbot     *template.Template
	getstats    *template.Template
	tourneyapi  *template.Template
	loggame     *template.Template
	trackersapi *template.Template
	teams       *template.Template
	apires      *template.Template
	user        *template.Template
	handler1    *template.Template
	roulette    *template.Template
	friends     *template.Template
	manage      *template.Template
}

func parsetemplates() {
	tmpl.shopbot = template.Must(template.ParseFiles("static/html/shopbot.html"))
	tmpl.getstats = template.Must(template.ParseFiles("static/html/stats.html"))
	tmpl.tourneyapi = template.Must(template.ParseFiles("static/html/tourneyapi.html"))
	tmpl.loggame = template.Must(template.ParseFiles("static/html/loggame.html"))
	tmpl.trackersapi = template.Must(template.ParseFiles("static/html/trackersapi.html"))
	tmpl.teams = template.Must(template.ParseFiles("static/html/teams.html"))
	tmpl.apires = template.Must(template.ParseFiles("static/html/apires.html"))
	tmpl.user = template.Must(template.ParseFiles("static/html/user.html"))
	tmpl.handler1 = template.Must(template.ParseFiles("static/html/home.html"))
	tmpl.roulette = template.Must(template.ParseFiles("static/html/roulette.html"))
	tmpl.friends = template.Must(template.ParseFiles("static/html/friends.html"))
	tmpl.manage = template.Must(template.ParseFiles("static/html/manage.html"))
}
