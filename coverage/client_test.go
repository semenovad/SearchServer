package main

import (
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

/*
	COVER
	go test -coverprofile=cover.out
	go tool cover -html=cover.out -o cover.html
*/
var (
	ts = httptest.NewServer(http.HandlerFunc(SearchServer))
)

const (
	token = "AccessToken"
)

type Result struct {
	Response *SearchResponse
	Error    error
}

type TestCase struct {
	Request SearchRequest
	Result  Result
}

func TestGreaterLimit(t *testing.T) {
	maximum := 25
	test := TestCase{
		Request: SearchRequest{
			Limit: maximum + 1,
		},
	}

	s := &SearchClient{
		token,
		ts.URL,
	}
	result, err := s.FindUsers(test.Request)

	if err != nil {
		t.Errorf("expected nil, got error")
	}
	if len(result.Users) > maximum {
		t.Errorf("wrong result, expected %v, got %#v", maximum, len(result.Users))
	}
}

type TestCaseServer struct {
	function func(w http.ResponseWriter, r *http.Request)
	result   string
}

func TestServerErr(t *testing.T) {
	tests := []TestCaseServer{
		{
			function: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusRequestTimeout)
				time.Sleep(1 * time.Second)
				return
			},
			result: "timeout",
		},
		{
			function: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				return
			},
			result: "SearchServer fatal error",
		},
		{
			function: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				io.WriteString(w, "StatusBadRequest")
				return
			},
			result: "cant unpack error json",
		},
		{
			function: func(w http.ResponseWriter, r *http.Request) {
				io.WriteString(w, "StatusBadRequest")
				return
			},
			result: "cant unpack result json",
		},
	}
	for caseNum, testItem := range tests {
		tmpTs := httptest.NewServer(http.HandlerFunc(testItem.function))
		s := &SearchClient{
			token,
			tmpTs.URL,
		}
		request := SearchRequest{}
		_, err := s.FindUsers(request)

		if err == nil {
			t.Errorf("[%v] expected timeout error, got nil", caseNum)
		}
		if !strings.Contains(err.Error(), testItem.result) {
			t.Errorf("[%v] wrong result, got %#v", caseNum, err.Error())
		}

		tmpTs.Close()
	}
}
func TestUnknownErrorBadAccess(t *testing.T) {
	tests := []SearchClient{
		{
			token,
			"",
		},
		{
			"bad",
			ts.URL,
		},
	}

	for caseNum, testItem := range tests {
		s := &testItem
		request := SearchRequest{}
		_, err := s.FindUsers(request)

		if err == nil {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !strings.Contains(err.Error(), "unknown error") && !strings.Contains(err.Error(), "Bad AccessToken") {
			t.Errorf("[%d] wrong result, got %#v", caseNum, err.Error())
		}
	}
}

type TestCaseString struct {
	request string
	result  string
}

func TestServerStatusBadRequest(t *testing.T) {
	tests := []TestCaseString{
		{
			request: "limit=&offset=0&query=&order_field=&order_by=0",
			result:  "no limit in request",
		},
		{
			request: "limit=1&offset=&query=&order_field=&order_by=0",
			result:  "no offset in request",
		},
		{
			request: "limit=1&offset=0&query=&order_field=&order_by=\"\"",
			result:  "no order_by in request",
		},
	}

	for caseNum, testItem := range tests {
		resp, _ := http.Get(ts.URL + "?" + testItem.request)
		body, _ := ioutil.ReadAll(resp.Body)
		errResp := SearchErrorResponse{}
		_ = json.Unmarshal(body, &errResp)
		if errResp.Error != testItem.result {
			t.Errorf("[%d] wrong result, got %#v", caseNum, errResp.Error)
		}
	}
}

func TestStatusBadRequestErrorBadOrderField(t *testing.T) {
	tests := []TestCase{
		{
			Request: SearchRequest{
				OrderBy:    1,
				OrderField: "N",
			},
			Result: Result{
				nil,
				errors.New("OrderFeld"),
			},
		},
		{
			Request: SearchRequest{
				OrderBy:    -1,
				OrderField: "N",
			},
			Result: Result{
				nil,
				errors.New("OrderFeld"),
			},
		},
		{
			Request: SearchRequest{
				OrderBy: 2,
			},
			Result: Result{
				nil,
				errors.New("unknown bad request error"),
			},
		},
		{
			Request: SearchRequest{
				Limit: -1,
			},
			Result: Result{
				nil,
				errors.New("limit must be > 0"),
			},
		},
		{
			Request: SearchRequest{
				Offset: -1,
			},
			Result: Result{
				nil,
				errors.New("offset must be > 0"),
			},
		},
	}

	for caseNum, testItem := range tests {
		s := &SearchClient{
			token,
			ts.URL,
		}
		_, err := s.FindUsers(testItem.Request)

		if err == nil {
			t.Errorf("[%d] expected error, got nil", caseNum)
		}
		if !strings.Contains(err.Error(), testItem.Result.Error.Error()) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, testItem.Result.Error.Error(), err.Error())
		}
	}
}

func TestOrderField(t *testing.T) {
	tests := []TestCase{
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    1,
				OrderField: "Id",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     0,
							Name:   "Boyd Wolf",
							Age:    22,
							About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
							Gender: "male",
						},
						User{
							Id:     1,
							Name:   "Hilda Mayer",
							Age:    21,
							About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
							Gender: "female",
						},
						User{
							Id:     2,
							Name:   "Brooks Aguilar",
							Age:    25,
							About:  "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    1,
				OrderField: "",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     15,
							Name:   "Allison Valdez",
							Age:    21,
							About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
							Gender: "male",
						},
						User{
							Id:     16,
							Name:   "Annie Osborn",
							Age:    35,
							About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
							Gender: "female",
						},
						User{
							Id:     19,
							Name:   "Bell Bauer",
							Age:    26,
							About:  "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    1,
				OrderField: "Name",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     15,
							Name:   "Allison Valdez",
							Age:    21,
							About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
							Gender: "male",
						},
						User{
							Id:     16,
							Name:   "Annie Osborn",
							Age:    35,
							About:  "Consequat fugiat veniam commodo nisi nostrud culpa pariatur. Aliquip velit adipisicing dolor et nostrud. Eu nostrud officia velit eiusmod ullamco duis eiusmod ad non do quis.\n",
							Gender: "female",
						},
						User{
							Id:     19,
							Name:   "Bell Bauer",
							Age:    26,
							About:  "Nulla voluptate nostrud nostrud do ut tempor et quis non aliqua cillum in duis. Sit ipsum sit ut non proident exercitation. Quis consequat laboris deserunt adipisicing eiusmod non cillum magna.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    1,
				OrderField: "Age",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     1,
							Name:   "Hilda Mayer",
							Age:    21,
							About:  "Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n",
							Gender: "female",
						},
						User{
							Id:     15,
							Name:   "Allison Valdez",
							Age:    21,
							About:  "Labore excepteur voluptate velit occaecat est nisi minim. Laborum ea et irure nostrud enim sit incididunt reprehenderit id est nostrud eu. Ullamco sint nisi voluptate cillum nostrud aliquip et minim. Enim duis esse do aute qui officia ipsum ut occaecat deserunt. Pariatur pariatur nisi do ad dolore reprehenderit et et enim esse dolor qui. Excepteur ullamco adipisicing qui adipisicing tempor minim aliquip.\n",
							Gender: "male",
						},
						User{
							Id:     23,
							Name:   "Gates Spencer",
							Age:    21,
							About:  "Dolore magna magna commodo irure. Proident culpa nisi veniam excepteur sunt qui et laborum tempor. Qui proident Lorem commodo dolore ipsum.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    -1,
				OrderField: "Id",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     34,
							Name:   "Kane Sharp",
							Age:    34,
							About:  "Lorem proident sint minim anim commodo cillum. Eiusmod velit culpa commodo anim consectetur consectetur sint sint labore. Mollit consequat consectetur magna nulla veniam commodo eu ut et. Ut adipisicing qui ex consectetur officia sint ut fugiat ex velit cupidatat fugiat nisi non. Dolor minim mollit aliquip veniam nostrud. Magna eu aliqua Lorem aliquip.\n",
							Gender: "male",
						},
						User{
							Id:     33,
							Name:   "Twila Snow",
							Age:    36,
							About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
							Gender: "female",
						},
						User{
							Id:     32,
							Name:   "Christy Knapp",
							Age:    40,
							About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
							Gender: "female",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    -1,
				OrderField: "",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     13,
							Name:   "Whitley Davidson",
							Age:    40,
							About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
							Gender: "male",
						},
						User{
							Id:     33,
							Name:   "Twila Snow",
							Age:    36,
							About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
							Gender: "female",
						},
						User{
							Id:     18,
							Name:   "Terrell Hall",
							Age:    27,
							About:  "Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    -1,
				OrderField: "Name",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     13,
							Name:   "Whitley Davidson",
							Age:    40,
							About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
							Gender: "male",
						},
						User{
							Id:     33,
							Name:   "Twila Snow",
							Age:    36,
							About:  "Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n",
							Gender: "female",
						},
						User{
							Id:     18,
							Name:   "Terrell Hall",
							Age:    27,
							About:  "Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      3,
				OrderBy:    -1,
				OrderField: "Age",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     13,
							Name:   "Whitley Davidson",
							Age:    40,
							About:  "Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n",
							Gender: "male",
						},
						User{
							Id:     32,
							Name:   "Christy Knapp",
							Age:    40,
							About:  "Incididunt culpa dolore laborum cupidatat consequat. Aliquip cupidatat pariatur sit consectetur laboris labore anim labore. Est sint ut ipsum dolor ipsum nisi tempor in tempor aliqua. Aliquip labore cillum est consequat anim officia non reprehenderit ex duis elit. Amet aliqua eu ad velit incididunt ad ut magna. Culpa dolore qui anim consequat commodo aute.\n",
							Gender: "female",
						},
						User{
							Id:     6,
							Name:   "Jennings Mays",
							Age:    39,
							About:  "Veniam consectetur non non aliquip exercitation quis qui. Aliquip duis ut ad commodo consequat ipsum cupidatat id anim voluptate deserunt enim laboris. Sunt nostrud voluptate do est tempor esse anim pariatur. Ea do amet Lorem in mollit ipsum irure Lorem exercitation. Exercitation deserunt adipisicing nulla aute ex amet sint tempor incididunt magna. Quis et consectetur dolor nulla reprehenderit culpa laboris voluptate ut mollit. Qui ipsum nisi ullamco sit exercitation nisi magna fugiat anim consectetur officia.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      2,
				Query:      "B",
				OrderBy:    1,
				OrderField: "Id",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     0,
							Name:   "Boyd Wolf",
							Age:    22,
							About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
							Gender: "male",
						},
						User{
							Id:     2,
							Name:   "Brooks Aguilar",
							Age:    25,
							About:  "Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n",
							Gender: "male",
						},
					},
					NextPage: true,
				},
				nil,
			},
		},
		{
			Request: SearchRequest{
				Limit:      10,
				Query:      "Boyd Wolf",
				OrderBy:    1,
				OrderField: "Id",
			},
			Result: Result{
				&SearchResponse{
					Users: []User{
						User{
							Id:     0,
							Name:   "Boyd Wolf",
							Age:    22,
							About:  "Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n",
							Gender: "male",
						},
					},
					NextPage: false,
				},
				nil,
			},
		},
	}

	for caseNum, testItem := range tests {
		s := &SearchClient{
			token,
			ts.URL,
		}
		result, err := s.FindUsers(testItem.Request)

		if err != nil {
			t.Errorf("[%d] expected nil, got error: %#v", caseNum, err.Error())
		}
		if !reflect.DeepEqual(testItem.Result.Response, result) {
			t.Errorf("[%d] wrong result, expected %#v, got %#v", caseNum, testItem.Result.Response, result)
		}
	}
}

type fileName struct {
	name   string
	result string
}

func TestIoutilAndJSON(t *testing.T) {
	tests := []fileName{
		{
			name:   "",
			result: "no such file or directory",
		},
		{
			name:   "coverfile.html",
			result: "can't unpack result json",
		},
	}

	for caseNum, testItem := range tests {
		FileName = testItem.name
		s := &SearchClient{
			token,
			ts.URL,
		}
		request := SearchRequest{}
		_, err := s.FindUsers(request)

		if err == nil {
			t.Errorf("[%v] expected error, got nil", caseNum)
		}
		if !strings.Contains(err.Error(), testItem.result) {
			t.Errorf("[%v] wrong result, expected \"unknown bad request error\", got %#v", caseNum, err.Error())
		}
	}
}
