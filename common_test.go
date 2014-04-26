package binding

import (
	"mime/multipart"
	"net/http"
	"testing"
)

// These types are mostly contrived examples, but they're used
// across many test cases. The idea is to cover all the scenarios
// that this binding package might encounter in actual use.
type (

	// For basic test cases with a required field
	Post struct {
		Title   string `form:"title" json:"title" binding:"required"`
		Content string `form:"content" json:"content"`
	}

	// To be used as a nested struct (with a required field)
	Person struct {
		Name  string `json:"name" binding:"required"`
		Email string `json:"email"`
	}

	// For advanced test cases: multiple values, embedded
	// and nested structs, an ignored field, and single
	// and multiple file uploads
	BlogPost struct {
		Post
		Id          int                     `form:"id" binding:"required"` // JSON not specified here for test coverage
		Ignored     string                  `form:"-" json:"-"`
		Ratings     []int                   `form:"ratings" json:"ratings"`
		Author      Person                  `form:"author" json:"author"`
		Coauthor    *Person                 `form:"coauthor" json:"coauthor"`
		HeaderImage *multipart.FileHeader   `form:"headerImage"`
		Pictures    []*multipart.FileHeader `form:"pictures"`
	}
)

func (p Post) Validate(errs Errors, req *http.Request) Errors {
	if len(p.Title) < 10 {
		errs = append(errs, Error{
			FieldNames:     []string{"title"},
			Classification: "LengthError",
			Message:        "Life is too short",
		})
	}
	return errs
}

func TestEnsureNotPointer(t *testing.T) {
	shouldPanic := func() {
		defer func() {
			r := recover()
			if r == nil {
				t.Errorf("Should have panicked because argument is a pointer, but did NOT panic")
			}
		}()
		ensureNotPointer(&Post{})
	}

	shouldNotPanic := func() {
		defer func() {
			r := recover()
			if r != nil {
				t.Errorf("Should NOT have panicked because argument is not a pointer, but did panic")
			}
		}()
		ensureNotPointer(Post{})
	}

	shouldPanic()
	shouldNotPanic()
}

const (
	testRoute = "/test"
)
