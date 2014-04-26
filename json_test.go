package binding

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-martini/martini"
)

var jsonTestCases = []jsonTestCase{
	{
		description:   "Happy path",
		shouldSucceed: true,
		method:        "POST",
		payload:       `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:   jsonContentType,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Nil payload should be handled gracefully",
		shouldSucceed: false,
		method:        "POST",
		payload:       `-nil-`,
		contentType:   jsonContentType,
		expected:      Post{},
	},
	{
		description:   "Empty payload",
		shouldSucceed: false,
		method:        "POST",
		payload:       ``,
		contentType:   jsonContentType,
		expected:      Post{},
	},
	{
		description:   "Empty content type",
		shouldSucceed: true,
		method:        "POST",
		payload:       `{"title": "Glorious Post Title", "content": "Lorem ipsum dolor sit amet"}`,
		contentType:   ``,
		expected:      Post{Title: "Glorious Post Title", Content: "Lorem ipsum dolor sit amet"},
	},
	{
		description:   "Malformed JSON",
		shouldSucceed: false,
		method:        "POST",
		payload:       `{"title":"foo"`,
		contentType:   ``,
		expected:      Post{},
	},
}

func TestJson(t *testing.T) {
	for _, testCase := range jsonTestCases {
		performJsonTest(t, testCase)
	}
}

func performJsonTest(t *testing.T, testCase jsonTestCase) {
	var payload io.Reader
	httpRecorder := httptest.NewRecorder()
	m := martini.Classic()

	m.Post(testRoute, Json(Post{}), func(actual Post, errs Errors) {
		if testCase.shouldSucceed && len(errs) > 0 {
			t.Errorf("'%s' should have succeeded, but there were errors (%d):\n%+v",
				testCase.description, len(errs), errs)
		}
		expString := fmt.Sprintf("%+v", testCase.expected)
		actString := fmt.Sprintf("%+v", actual)
		if actString != expString {
			t.Errorf("'%s': expected\n'%s'\nbut got\n'%s'",
				testCase.description, expString, actString)
		}
	})

	if testCase.payload == "-nil-" {
		payload = nil
	} else {
		payload = strings.NewReader(testCase.payload)
	}

	req, err := http.NewRequest(testCase.method, testRoute, payload)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", testCase.contentType)

	m.ServeHTTP(httpRecorder, req)

	switch httpRecorder.Code {
	case http.StatusNotFound:
		panic("Routing is messed up in test fixture (got 404): check method and path")
	case http.StatusInternalServerError:
		panic("Something bad happened on '" + testCase.description + "'")
	}
}

type (
	jsonTestCase struct {
		description   string
		shouldSucceed bool
		method        string
		payload       string
		contentType   string
		expected      Post
	}
)
