package post

import (
	"errors"
	"regexp"

	"github.com/asaskevich/govalidator"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"reddit/pkg/comment"
	"reddit/pkg/user"
	"reddit/pkg/vote"
)

var (
	ErrNoPost    = errors.New("no post found")
	ErrNoAccess  = errors.New("forbidden action")
	ErrNoComment = errors.New("no comment found")
)

type PostRepo interface {
	GetAll() ([]*Post, error)
	AddPost(post *Post, author *user.User) (*Post, error)
	GetPostByCategory(category string) ([]*Post, error)
	GetPostByID(ID string) (*Post, error)
	AddComment(commentBody string, author *user.User, postID string) (*Post, error)
	DeleteComment(userID, postID string, commentID string) (*Post, error)
	UpVote(postID string, userID string) (*Post, error)
	DownVote(postID string, userID string) (*Post, error)
	UnVote(postID string, userID string) (*Post, error)
	DeletePost(userID, postID string) (bool, error)
	GetPostsByUserID(userName string) ([]*Post, error)
}

type Post struct {
	Score            int                `json:"score" bson:"score"`
	Views            int                `json:"views" bson:"views"`
	Type             string             `json:"type" bson:"type" valid:"required,in(text|link)"`
	Title            string             `json:"title" bson:"title" valid:"required,length(1|100)"`
	URL              string             `json:"url,omitempty" bson:"url" valid:"url"`
	Author           *user.User         `json:"author" bson:"author"`
	Category         string             `json:"category" bson:"category" valid:"required,length(1|300)"`
	Text             string             `json:"text,omitempty" bson:"text"`
	Votes            []*vote.Vote       `json:"votes" bson:"votes"`
	Comments         []*comment.Comment `json:"comments" bson:"comments"`
	Created          string             `json:"created" bson:"created"`
	UpvotePercentage int                `json:"upvotePercentage" bson:"upvotePercentage"`
	ID               primitive.ObjectID `json:"id" bson:"_id"`
}

func init() {
	govalidator.CustomTypeTagMap.Set("url", govalidator.CustomTypeValidator(func(i interface{}, o interface{}) bool {
		subject, ok := i.(string)
		if !ok {
			return false
		}
		urlPattern := `^(http|https):\/\/[a-zA-Z0-9.-]+(\.[a-zA-Z]{2,}){1,}([\w\W]*)$`
		re := regexp.MustCompile(urlPattern)
		return re.MatchString(subject)
	}))

}

func (p *Post) Validate() []string {
	_, err := govalidator.ValidateStruct(p)
	validationErrors := make([]string, 0)
	if err == nil {
		return validationErrors
	}
	if allErrs, ok := err.(govalidator.Errors); ok {
		for _, fld := range allErrs {
			validationErrors = append(validationErrors, fld.Error())
		}
	}
	if p.Type == "text" {
		if p.Text == "" {
			validationErrors = append(validationErrors, "text field required")

		}
	} else if p.URL == "" {
		validationErrors = append(validationErrors, "url field required")
	}
	return validationErrors

}
