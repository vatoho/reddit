package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
	"reddit/pkg/response"

	"github.com/gorilla/mux"
	"reddit/pkg/comment"
	"reddit/pkg/middleware"
	"reddit/pkg/post"
	"reddit/pkg/user"
)

type PostHandler struct {
	PostRepo post.PostRepo
	Logger   *zap.SugaredLogger
}

func (ph *PostHandler) List(w http.ResponseWriter, _ *http.Request) {
	posts, err := ph.PostRepo.GetAll()
	if err != nil {
		errText := fmt.Sprintf(`{"message": "can not get posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	postsJSON, err := json.Marshal(posts)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	response.WriteResponse(ph.Logger, w, postsJSON, http.StatusOK)
}

func (ph *PostHandler) NewPost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	author, ok := ctx.Value(middleware.MyUserKey).(*user.User)
	if !ok {
		response.WriteResponse(ph.Logger, w, []byte(`{"message": "can not cast context value to user"}`), http.StatusInternalServerError)
		return
	}
	postFromForm := &post.Post{}
	rBody, err := io.ReadAll(r.Body)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in reading request body: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(rBody, postFromForm)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in json decoding of post form: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	ph.Logger.Infof("postForm %v", postFromForm)

	if validationErrors := postFromForm.Validate(); len(validationErrors) != 0 {
		var errorsJSON []byte
		errorsJSON, err = json.Marshal(validationErrors)
		if err != nil {
			errText := fmt.Sprintf(`{"message": "error in json coding of validation errors of post: %s"}`, err)
			response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
			return
		}
		response.WriteResponse(ph.Logger, w, errorsJSON, http.StatusUnprocessableEntity)
		return
	}

	addedPost, err := ph.PostRepo.AddPost(postFromForm, author)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in adding post: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	newPostJSON, err := json.Marshal(addedPost)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	ph.Logger.Infof("new post with id %s", addedPost.ID)
	response.WriteResponse(ph.Logger, w, newPostJSON, http.StatusCreated)

}

func (ph *PostHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	category := vars["CATEGORY_NAME"]
	posts, err := ph.PostRepo.GetPostByCategory(category)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "can not get posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	postsJSON, err := json.Marshal(posts)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	response.WriteResponse(ph.Logger, w, postsJSON, http.StatusOK)
}

func (ph *PostHandler) GetPostInfo(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	curPost, err := ph.PostRepo.GetPostByID(postID)

	if errors.Is(err, post.ErrNoPost) {
		errText := fmt.Sprintf(`{"message": "there is no post with id %s"}`, postID)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusNotFound)
		return
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "can not get post: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	postsJSON, err := json.Marshal(curPost)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	response.WriteResponse(ph.Logger, w, postsJSON, http.StatusOK)

}

func (ph *PostHandler) NewComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	ctx := r.Context()
	author, ok := ctx.Value(middleware.MyUserKey).(*user.User)
	if !ok {
		response.WriteResponse(ph.Logger, w, []byte(`{"message": "can not cast context value to user"}`), http.StatusInternalServerError)
		return
	}
	commentFromForm := &comment.CommentForm{}
	rBody, err := io.ReadAll(r.Body)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in reading request body: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(rBody, commentFromForm)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in decoding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	ph.Logger.Infof("comment form %v", commentFromForm)

	if validationErrors := commentFromForm.Validate(); len(validationErrors) != 0 {
		var errorsJSON []byte
		errorsJSON, err = json.Marshal(validationErrors)
		if err != nil {
			errText := fmt.Sprintf(`{"message": "error in json decoding: %s"}`, err)
			response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
			return
		}
		response.WriteResponse(ph.Logger, w, errorsJSON, http.StatusUnprocessableEntity)
		return
	}
	myPost, err := ph.PostRepo.AddComment(commentFromForm.Body, author, postID)
	if errors.Is(err, post.ErrNoPost) {
		errText := fmt.Sprintf(`{"message": "there is no post with id %s"}`, postID)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusNotFound)
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in adding new comment: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	postJSON, err := json.Marshal(myPost)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	ph.Logger.Infof("new comment created")
	response.WriteResponse(ph.Logger, w, postJSON, http.StatusCreated)
}

func (ph *PostHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	commentID := vars["COMMENT_ID"]
	ctx := r.Context()
	currentUser, ok := ctx.Value(middleware.MyUserKey).(*user.User)
	if !ok {
		response.WriteResponse(ph.Logger, w, []byte(`{"message": "can not cast context value to user"}`), http.StatusInternalServerError)
		return
	}
	myPost, err := ph.PostRepo.DeleteComment(currentUser.ID, postID, commentID)
	if errors.Is(err, post.ErrNoAccess) {
		errText := fmt.Sprintf(`{"message": "forbidden for this user: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusForbidden)
		return
	}
	if errors.Is(err, post.ErrNoPost) {
		errText := fmt.Sprintf(`{"message": "there is no post with id %s"}`, postID)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusNotFound)
		return
	}
	if errors.Is(err, post.ErrNoComment) {
		errText := fmt.Sprintf(`{"message": "there is no comment with id %s"}`, postID)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusNotFound)
		return
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in comment deletion: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	postJSON, err := json.Marshal(myPost)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	ph.Logger.Infof("comment deleted")
	response.WriteResponse(ph.Logger, w, postJSON, http.StatusOK)

}

func (ph *PostHandler) MakeVote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	ctx := r.Context()
	curUser, ok := ctx.Value(middleware.MyUserKey).(*user.User)
	if !ok {
		response.WriteResponse(ph.Logger, w, []byte(`{"message": "can not cast context value to user"}`), http.StatusInternalServerError)
		return
	}
	segmentsURL := strings.Split(r.URL.Path, "/")
	voteAction := segmentsURL[len(segmentsURL)-1]
	var myPost *post.Post
	var err error
	switch voteAction {
	case "upvote":
		myPost, err = ph.PostRepo.UpVote(postID, curUser.ID)
	case "downvote":
		myPost, err = ph.PostRepo.DownVote(postID, curUser.ID)
	default:
		myPost, err = ph.PostRepo.UnVote(postID, curUser.ID)
	}
	if errors.Is(err, post.ErrNoPost) {
		errText := fmt.Sprintf(`{"message": "there is no post with id %s"}`, postID)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusNotFound)
		return
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in post %s"}`, voteAction)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	postJSON, err := json.Marshal(myPost)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	ph.Logger.Infof("vote added/deleted")
	response.WriteResponse(ph.Logger, w, postJSON, http.StatusOK)

}

func (ph *PostHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postID := vars["POST_ID"]
	ctx := r.Context()
	currentUser, ok := ctx.Value(middleware.MyUserKey).(*user.User)
	if !ok {
		response.WriteResponse(ph.Logger, w, []byte(`{"message": "can not cast context value to user"}`), http.StatusInternalServerError)
		return
	}
	isDeleted, err := ph.PostRepo.DeletePost(currentUser.ID, postID)

	if errors.Is(err, post.ErrNoPost) {
		errText := fmt.Sprintf(`{"message": "there is no post with id %s"}`, postID)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusNotFound)
		return
	}
	if errors.Is(err, post.ErrNoAccess) {
		errText := fmt.Sprintf(`{"message": "forbidden for this user: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusForbidden)
		return
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in post deletion: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	if isDeleted {
		response.WriteResponse(ph.Logger, w, []byte(`{"message": "success"}`), http.StatusOK)
		return
	}
	response.WriteResponse(ph.Logger, w, []byte(`{"message": "fail"}`), http.StatusUnprocessableEntity)
}

func (ph *PostHandler) ListByUserLogin(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userLogin := vars["USER_LOGIN"]
	posts, err := ph.PostRepo.GetPostsByUserID(userLogin)
	if errors.Is(err, user.ErrNoUser) {
		errText := fmt.Sprintf(`{"message": "there is no user with username %s"}`, userLogin)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusNotFound)
	}
	if err != nil {
		errText := fmt.Sprintf(`{"message": "can not get posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	w.Header()
	postsJSON, err := json.Marshal(posts)
	if err != nil {
		errText := fmt.Sprintf(`{"message": "error in coding posts: %s"}`, err)
		response.WriteResponse(ph.Logger, w, []byte(errText), http.StatusInternalServerError)
		return
	}
	response.WriteResponse(ph.Logger, w, postsJSON, http.StatusOK)
}
