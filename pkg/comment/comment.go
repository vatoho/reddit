package comment

import (
	"github.com/asaskevich/govalidator"
	"reddit/pkg/user"
)

type Comment struct {
	Created string     `json:"created"`
	Author  *user.User `json:"author"`
	Body    string     `json:"body"`
	ID      string     `json:"id"`
}

func (c *CommentForm) Validate() []string {
	_, err := govalidator.ValidateStruct(c)
	if err == nil {
		return nil
	}
	validationErrors := make([]string, 0)
	if allErrs, ok := err.(govalidator.Errors); ok {
		for _, fld := range allErrs {
			validationErrors = append(validationErrors, fld.Error())
		}
	}
	return validationErrors
}

type CommentForm struct {
	Body string `json:"comment" valid:"required,length(1|1000)"`
}
