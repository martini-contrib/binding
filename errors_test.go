package binding

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFieldErrors(t *testing.T) {
	errsFieldRequired := NewErrors()
	errsFieldRequired.Fields["foo"] = RequireError
	performErrorsTest(t, errorTestCase{
		description: "Required field error",
		errors:      errsFieldRequired,
		expected: errorTestResult{
			statusCode:  StatusUnprocessableEntity,
			contentType: jsonContentType,
			body:        `{"overall":{},"fields":{"foo":"Required"}}`,
		},
	})

	errsFieldCustom := NewErrors()
	errsFieldCustom.Fields["bar"] = "foo"
	performErrorsTest(t, errorTestCase{
		description: "Custom field error",
		errors:      errsFieldCustom,
		expected: errorTestResult{
			statusCode:  StatusUnprocessableEntity,
			contentType: jsonContentType,
			body:        `{"overall":{},"fields":{"bar":"foo"}}`,
		},
	})
}

func TestOverallErrors(t *testing.T) {
	errsDeserialization := NewErrors()
	errsDeserialization.Overall[DeserializationError] = "Foo parser error"
	performErrorsTest(t, errorTestCase{
		description: "Deserialization error",
		errors:      errsDeserialization,
		expected: errorTestResult{
			statusCode:  http.StatusBadRequest,
			contentType: jsonContentType,
			body:        `{"overall":{"DeserializationError":"Foo parser error"},"fields":{}}`,
		},
	})

	errsContentType := NewErrors()
	errsContentType.Overall[ContentTypeError] = "Empty Content-Type"
	performErrorsTest(t, errorTestCase{
		description: "Content-Type error",
		errors:      errsContentType,
		expected: errorTestResult{
			statusCode:  http.StatusUnsupportedMediaType,
			contentType: jsonContentType,
			body:        `{"overall":{"ContentTypeError":"Empty Content-Type"},"fields":{}}`,
		},
	})

	errsCustomOverall := NewErrors()
	errsCustomOverall.Overall["BadHeader"] = "Some message here"
	performErrorsTest(t, errorTestCase{
		description: "Custom overall error",
		errors:      errsCustomOverall,
		expected: errorTestResult{
			statusCode:  StatusUnprocessableEntity,
			contentType: jsonContentType,
			body:        `{"overall":{"BadHeader":"Some message here"},"fields":{}}`,
		},
	})
}

func TestNoErrors(t *testing.T) {
	errsNone := NewErrors()
	performErrorsTest(t, errorTestCase{
		description: "No errors",
		errors:      errsNone,
		expected: errorTestResult{
			statusCode:  http.StatusOK,
			contentType: "",
			body:        ``,
		},
	})
}

func TestErrorCombine(t *testing.T) {
	errs1, errs2 := NewErrors(), NewErrors()
	errs1.Overall["foo1"] = "foo1"
	errs1.Fields["bar1"] = "bar1"
	errs2.Overall["foo2"] = "foo2"
	errs2.Fields["bar2"] = "bar2"

	errs1.combine(*errs2)

	actual := fmt.Sprintf("%+v", errs1)
	expected := `&{Overall:map[foo1:foo1 foo2:foo2] Fields:map[bar1:bar1 bar2:bar2]}`

	if actual != expected {
		t.Errorf("Expected errors to combine like so: '%s' - but got '%s' instead",
			expected, actual)
	}
}

func TestErrorCount(t *testing.T) {
	errs := NewErrors()
	if errs.Count() != 0 {
		t.Errorf("Expected error count to be 0, but it was %d", errs.Count())
	}
	errs.Overall["foo"] = "foo"
	if errs.Count() != 1 {
		t.Errorf("Expected error count to be 1, but it was %d", errs.Count())
	}
	errs.Overall["bar"] = "bar"
	errs.Fields["baz"] = "baz"
	if errs.Count() != 3 {
		t.Errorf("Expected error count to be 3, but it was %d", errs.Count())
	}
}

func performErrorsTest(t *testing.T, testCase errorTestCase) {
	httpRecorder := httptest.NewRecorder()

	// Executes the test
	ErrorHandler(*testCase.errors, httpRecorder)

	actualBody, _ := ioutil.ReadAll(httpRecorder.Body)
	actualContentType := httpRecorder.Header().Get("Content-Type")

	if httpRecorder.Code != testCase.expected.statusCode {
		t.Errorf("For '%s': expected status code %d but got %d instead",
			testCase.description, testCase.expected.statusCode, httpRecorder.Code)
	}
	if actualContentType != testCase.expected.contentType {
		t.Errorf("For '%s': expected content-type '%s' but got '%s' instead",
			testCase.description, testCase.expected.contentType, actualContentType)
	}
	if string(actualBody) != testCase.expected.body {
		t.Errorf("For '%s': expected body to be '%s' but got '%s' instead",
			testCase.description, testCase.expected.body, actualBody)
	}
}

type (
	errorTestCase struct {
		description string
		errors      *Errors
		expected    errorTestResult
	}

	errorTestResult struct {
		statusCode  int
		contentType string
		body        string
	}
)
