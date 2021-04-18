package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type row struct {
	ID            int    `xml:"id"`
	GUID          string `xml:"guid"`
	IsActive      bool   `xml:"isActive"`
	Balance       string `xml:"balance"`
	Picture       string `xml:"picture"`
	Age           int    `xml:"age"`
	EyeColor      string `xml:"eyeColor"`
	FirstName     string `xml:"first_name"`
	LastName      string `xml:"last_name"`
	Gender        string `xml:"gender"`
	Company       string `xml:"company"`
	Email         string `xml:"email"`
	Phone         string `xml:"phone"`
	Address       string `xml:"address"`
	About         string `xml:"about"`
	Registered    string `xml:"registered"`
	FavoriteFruit string `xml:"favoriteFruit"`
}

type root struct {
	RowMas []row `xml:"row"`
}

//FileName is cool
var FileName = "dataset.xml"

//SearchServer is cool
func SearchServer(w http.ResponseWriter, r *http.Request) {
	xmlData, err := ioutil.ReadFile(FileName)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
		io.WriteString(w, `{"error": "no such file or directory"}`)
		return
	}

	t := r.Header.Get("AccessToken")
	if t == "bad" {
		w.WriteHeader(http.StatusUnauthorized) //StatusUnauthorized
		io.WriteString(w, "Bad AccessToken")
		return
	}

	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
		io.WriteString(w, `{"error": "no limit in request"}`)
		return
	}

	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
		io.WriteString(w, `{"error": "no offset in request"}`)
		return
	}

	orderBy, err := strconv.Atoi(r.FormValue("order_by"))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
		io.WriteString(w, `{"error": "no order_by in request"}`)
		return
	}

	query := r.FormValue("query")
	orderField := r.FormValue("order_field")

	req := SearchRequest{
		limit, offset, query, orderField, orderBy,
	}

	xmlUsers := new(root)
	err = xml.Unmarshal(xmlData, &xmlUsers)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
		io.WriteString(w, `{"error": "can't unpack result json"}`)
		return
	}

	var user User
	var users []User
	for _, row := range xmlUsers.RowMas {
		name := row.FirstName + " " + row.LastName
		user = User{row.ID, name, row.Age, row.About, row.Gender}
		if strings.Contains(row.About, req.Query) || strings.Contains(name, req.Query) || req.Query == "" {
			users = append(users, user)
		}
	}

	switch req.OrderBy {
	case 0:
	case 1:
		{
			switch req.OrderField {
			case "Id":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Id < users[j].Id })
				}
			case "":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Name < users[j].Name })
				}
			case "Name":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Name < users[j].Name })
				}
			case "Age":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Age < users[j].Age })
				}
			default:
				{
					w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
					io.WriteString(w, `{"error": "ErrorBadOrderField"}`)
					return
				}
			}
		}
	case -1:
		{
			switch req.OrderField {
			case "Id":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Id > users[j].Id })
				}
			case "":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Name > users[j].Name })
				}
			case "Name":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Name > users[j].Name })
				}
			case "Age":
				{
					sort.SliceStable(users, func(i, j int) bool { return users[i].Age > users[j].Age })
				}
			default:
				{
					w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
					io.WriteString(w, `{"error": "ErrorBadOrderField"}`)
					return
				}
			}
		}
	default:
		{
			w.WriteHeader(http.StatusBadRequest) //StatusBadRequest
			io.WriteString(w, `{"error": "have no such sort parameter"}`)
			return
		}

	}

	if len(users) >= req.Limit+req.Offset {
		users = users[req.Offset:(req.Limit + req.Offset)]
	}

	usersToJSON, err := json.Marshal(users)
	if err != nil {
		io.WriteString(w, `{"error": "can't Marshal users to usersToJSON"}`)
		return
	}
	w.Write(usersToJSON)
}

func main() {}
