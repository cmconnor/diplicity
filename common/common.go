package common

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

const (
	regular = iota
	nilCache
)

type Variant struct {
	Id          string
	Name        string
	Translation string
}

var Variants = []Variant{
	Variant{
		Id:   "standard",
		Name: "Standard",
	},
}

var prefPattern = regexp.MustCompile("^([^\\s;]+)(;q=([\\d.]+))?$")

func MustParseFloat64(s string) (result float64) {
	var err error
	if result, err = strconv.ParseFloat(s, 64); err != nil {
		panic(err)
	}
	return
}

func UserRoot(c appengine.Context, email string) *datastore.Key {
	return datastore.NewKey(c, "Root", email, 0, nil)
}

func MustParseInt64(s string) (result int64) {
	var err error
	if result, err = strconv.ParseInt(s, 10, 64); err != nil {
		panic(err)
	}
	return
}

func MustParseInt(s string) (result int) {
	var err error
	if result, err = strconv.Atoi(s); err != nil {
		panic(err)
	}
	return
}

func AssertOkError(err error) {
	if err != nil {
		if merr, ok := err.(appengine.MultiError); ok {
			for _, serr := range merr {
				if serr != nil {
					if _, ok := serr.(*datastore.ErrFieldMismatch); !ok {
						panic(err)
					}
				}
			}
		} else if _, ok := err.(*datastore.ErrFieldMismatch); !ok {
			panic(err)
		}
	}
}

func SetContentType(w http.ResponseWriter, typ string) {
	w.Header().Set("Content-Type", typ)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}

func MustEncodeJSON(w io.Writer, i interface{}) {
	if err := json.NewEncoder(w).Encode(i); err != nil {
		panic(err)
	}
}

func MostAccepted(r *http.Request, def, name string) string {
	bestValue := def
	var bestScore float64 = -1
	var score float64
	for _, pref := range strings.Split(r.Header.Get(name), ",") {
		if match := prefPattern.FindStringSubmatch(pref); match != nil {
			score = 1
			if match[3] != "" {
				score = MustParseFloat64(match[3])
			}
			if score > bestScore {
				bestScore = score
				bestValue = match[1]
			}
		}
	}
	return bestValue
}

func HostURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%v://%v/reload", scheme, r.Host)
}

func Login(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	loginUrl, err := user.LoginURL(data.Context, HostURL(data.Request))
	if err != nil {
		panic(err)
	}
	data.Response.Header().Set("Location", loginUrl)
	data.Response.WriteHeader(302)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	logoutUrl, err := user.LogoutURL(data.Context, HostURL(data.Request))
	if err != nil {
		panic(err)
	}
	data.Response.Header().Set("Location", logoutUrl)
	data.Response.WriteHeader(302)
}

type jsonUser struct {
	Admin bool   `json:"admin"`
	Email string `json:"email"`
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	data := GetRequestData(w, r)
	SetContentType(w, "application/json; charset=UTF-8")
	if data.User == nil {
		MustEncodeJSON(w, jsonUser{})
	} else {
		MustEncodeJSON(w, jsonUser{
			Admin: data.User.Admin,
			Email: data.User.Email,
		})
	}
}
