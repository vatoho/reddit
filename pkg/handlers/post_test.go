package handlers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	"reddit/pkg/comment"
	"reddit/pkg/middleware"
	"reddit/pkg/post"
	"reddit/pkg/user"
	"reddit/pkg/vote"
)

func TestPostHandlerList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	testRepo.EXPECT().GetAll().Return(nil, fmt.Errorf("error"))
	request := httptest.NewRequest(http.MethodGet, "/api/posts/", nil)
	respWriter := httptest.NewRecorder()
	testHandler.List(respWriter, request)
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return
	}
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}

	posts := []*post.Post{
		{
			Score: 1,
			Views: 0,
			Type:  "text",
			Title: "fef",
			Author: &user.User{
				ID:       "310ca263",
				Username: "hhhhhhhh",
			},
			Category: "programming",
			Text:     "rferfer",
			Votes: []*vote.Vote{
				{
					Value:  1,
					UserID: "310ca263",
				},
			},

			Comments:         []*comment.Comment{},
			Created:          "2023-11-11T14:22:11.695Z",
			UpvotePercentage: 100,
			ID:               objID,
		},
	}
	testRepo.EXPECT().GetAll().Return(posts, nil)
	request = httptest.NewRequest(http.MethodGet, "/api/posts/", nil)
	respWriter = httptest.NewRecorder()
	testHandler.List(respWriter, request)
	resp = respWriter.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}

	expectedBodySuccess :=
		`[{"score":1,"views":0,"type":"text","title":"fef","author":{"id":"310ca263","username":"hhhhhhhh"},"category":"programming","text":"rferfer","votes":[{"vote":1,"user":"310ca263"}],"comments":[],"created":"2023-11-11T14:22:11.695Z","upvotePercentage":100,"id":"654f63e3a2414a2a554b6423"}]`

	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
		return
	}
	if string(body) != expectedBodySuccess {
		t.Errorf("wrond response body: \nexpected %s, \ngot      %s", expectedBodySuccess, string(body))
	}

}

func TestPostHandlerNewPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	//  в контексте лежит битый юзер
	request := httptest.NewRequest(http.MethodPost, "/api/posts", nil)
	ctx := request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, "bad user value")
	respWriter := httptest.NewRecorder()
	testHandler.NewPost(respWriter, request.WithContext(ctx))
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return
	}

	//  не получилось считать запрос
	request = httptest.NewRequest(http.MethodPost, "/api/posts", &errorReader{})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "fd3f43f3",
		Username: "rvfvryby",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewPost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status %d, got status %d", http.StatusBadRequest, resp.StatusCode)
	}

	//  не получилось сделать анмаршал запроса
	request = httptest.NewRequest(http.MethodPost, "/api/posts", strings.NewReader(`{""`))
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "fd3f43f3",
		Username: "rvfvryby",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewPost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusBadRequest, resp.StatusCode)
	}

	//  не прошел валидацию, не получается расшифровать ошибки валидации

	//  не прошел валидацию, ошибки расшифрованы
	request = httptest.NewRequest(http.MethodPost, "/api/posts",
		strings.NewReader(`{"type":"bad type"}`))
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "fd3f43f3",
		Username: "rvfvryby",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewPost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 422 {
		t.Errorf("expected status %d, got status %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}

	//  метод addpost вернул ошибку
	postToAdd := &post.Post{
		Category: "programming",
		Text:     "rferfer",
		Title:    "fef",
		Type:     "text",
	}
	authorOfPost := &user.User{
		ID:       "310ca263",
		Username: "hhhhhhhh",
	}

	testRepo.EXPECT().AddPost(postToAdd, authorOfPost).Return(nil, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodPost, "/api/posts",
		strings.NewReader(`{"category":"programming","text":"rferfer","title":"fef","type":"text"}`))
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "310ca263",
		Username: "hhhhhhhh",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewPost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
	}

	//  пост добавлен
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	createdPost := &post.Post{
		Score: 1,
		Views: 0,
		Type:  "text",
		Title: "fef",
		Author: &user.User{
			ID:       "310ca263",
			Username: "hhhhhhhh",
		},
		Category: "programming",
		Text:     "rferfer",
		Votes: []*vote.Vote{
			{
				Value:  1,
				UserID: "310ca263",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}

	testRepo.EXPECT().AddPost(postToAdd, authorOfPost).Return(createdPost, nil)
	request = httptest.NewRequest(http.MethodPost, "/api/posts",
		strings.NewReader(`{"category":"programming","text":"rferfer","title":"fef","type":"text"}`))
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "310ca263",
		Username: "hhhhhhhh",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewPost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected status %d, got status %d", http.StatusCreated, resp.StatusCode)
	}

}

type errorReader struct{}

func (er *errorReader) Read(_ []byte) (int, error) {
	return 0, io.ErrUnexpectedEOF
}

func TestPostHandlerListByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	// ошибка при поиске постов
	testRepo.EXPECT().GetPostByCategory("programming").Return(nil, fmt.Errorf("error"))
	request := httptest.NewRequest(http.MethodGet, "/api/posts/programming", nil)
	request = mux.SetURLVars(request, map[string]string{"CATEGORY_NAME": "programming"})

	respWriter := httptest.NewRecorder()
	testHandler.ListByCategory(respWriter, request)
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return
	}

	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}

	posts := []*post.Post{
		{
			Score: 1,
			Views: 0,
			Type:  "text",
			Title: "fef",
			Author: &user.User{
				ID:       "310ca263",
				Username: "hhhhhhhh",
			},
			Category: "programming",
			Text:     "rferfer",
			Votes: []*vote.Vote{
				{
					Value:  1,
					UserID: "310ca263",
				},
			},

			Comments:         []*comment.Comment{},
			Created:          "2023-11-11T14:22:11.695Z",
			UpvotePercentage: 100,
			ID:               objID,
		},
	}

	//  корректный ответ с постами
	testRepo.EXPECT().GetPostByCategory("programming").Return(posts, nil)
	request = httptest.NewRequest(http.MethodGet, "/api/posts/programming", nil)
	request = mux.SetURLVars(request, map[string]string{"CATEGORY_NAME": "programming"})
	respWriter = httptest.NewRecorder()
	testHandler.ListByCategory(respWriter, request)
	resp = respWriter.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}

	expectedBodySuccess :=
		`[{"score":1,"views":0,"type":"text","title":"fef","author":{"id":"310ca263","username":"hhhhhhhh"},"category":"programming","text":"rferfer","votes":[{"vote":1,"user":"310ca263"}],"comments":[],"created":"2023-11-11T14:22:11.695Z","upvotePercentage":100,"id":"654f63e3a2414a2a554b6423"}]`

	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
		return
	}
	if string(body) != expectedBodySuccess {
		t.Errorf("wrond response body: \nexpected %s, \ngot      %s", expectedBodySuccess, string(body))
	}

}

func TestPostHandlerGetPostInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	// пост не найден
	testRepo.EXPECT().GetPostByID("id_which_not_exists").Return(nil, post.ErrNoPost)
	request := httptest.NewRequest(http.MethodGet, "/api/posts/id_which_not_exists", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "id_which_not_exists"})

	respWriter := httptest.NewRecorder()
	testHandler.GetPostInfo(respWriter, request)
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return
	}

	//  какая то ошибка сервера при поиске поста
	testRepo.EXPECT().GetPostByID("hrebhrbfher").Return(nil, fmt.Errorf("internal error"))
	request = httptest.NewRequest(http.MethodGet, "/api/posts/hrebhrbfher", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "hrebhrbfher"})

	respWriter = httptest.NewRecorder()
	testHandler.GetPostInfo(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return
	}

	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}

	post := &post.Post{
		Score: 1,
		Views: 0,
		Type:  "text",
		Title: "fef",
		Author: &user.User{
			ID:       "310ca263",
			Username: "hhhhhhhh",
		},
		Category: "programming",
		Text:     "rferfer",
		Votes: []*vote.Vote{
			{
				Value:  1,
				UserID: "310ca263",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}

	// пост найден
	testRepo.EXPECT().GetPostByID("654f63e3a2414a2a554b6423").Return(post, nil)
	request = httptest.NewRequest(http.MethodGet, "/api/posts/654f63e3a2414a2a554b6423", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "654f63e3a2414a2a554b6423"})
	respWriter = httptest.NewRecorder()
	testHandler.GetPostInfo(respWriter, request)
	resp = respWriter.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}

	expectedBodySuccess :=
		`{"score":1,"views":0,"type":"text","title":"fef","author":{"id":"310ca263","username":"hhhhhhhh"},"category":"programming","text":"rferfer","votes":[{"vote":1,"user":"310ca263"}],"comments":[],"created":"2023-11-11T14:22:11.695Z","upvotePercentage":100,"id":"654f63e3a2414a2a554b6423"}`

	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
		return
	}
	if string(body) != expectedBodySuccess {
		t.Errorf("wrond response body: \nexpected %s, \ngot      %s", expectedBodySuccess, string(body))
	}
}

func TestPostHandlerNewComment(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	//  в контексте лежит битый юзер
	request := httptest.NewRequest(http.MethodPost, "/api/post/dwdwdwe", nil)
	ctx := request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, "bad user value")
	respWriter := httptest.NewRecorder()
	testHandler.NewComment(respWriter, request.WithContext(ctx))
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return
	}

	//  не получилось считать запрос
	request = httptest.NewRequest(http.MethodPost, "/api/post/efefefe", &errorReader{})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "fd3f43f3",
		Username: "rvfvryby",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 400 {
		t.Errorf("expected status %d, got status %d", http.StatusBadRequest, resp.StatusCode)
	}

	//  не получилось сделать анмаршал запроса
	request = httptest.NewRequest(http.MethodPost, "/api/post/bhefbhde", strings.NewReader(`{""`))
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "fd3f43f3",
		Username: "rvfvryby",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusBadRequest, resp.StatusCode)
	}

	//  не прошел валидацию, ошибки расшифрованы
	request = httptest.NewRequest(http.MethodPost, "/api/post/frffef",
		strings.NewReader(`{}`))
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, &user.User{
		ID:       "fd3f43f3",
		Username: "rvfvryby",
	})
	respWriter = httptest.NewRecorder()
	testHandler.NewComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}

	if resp.StatusCode != 422 {
		t.Errorf("expected status %d, got status %d", http.StatusUnprocessableEntity, resp.StatusCode)
	}

	//  поста к которому хотят сделать коммент не существует
	authorOfPost := &user.User{
		ID:       "310ca263",
		Username: "hhhhhhhh",
	}

	testRepo.EXPECT().AddComment("some comment", authorOfPost, "not_exist_post").Return(nil, post.ErrNoPost)
	request = httptest.NewRequest(http.MethodPost, "/api/post/not_exist_post",
		strings.NewReader(`{"comment":"some comment"}`))
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "not_exist_post"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, authorOfPost)
	respWriter = httptest.NewRecorder()
	testHandler.NewComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return

	}

	//  неизвестная ошибка при добавлении коммента
	testRepo.EXPECT().AddComment("some comment", authorOfPost, "some_post").Return(nil, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodPost, "/api/post/some_post",
		strings.NewReader(`{"comment":"some comment"}`))
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "some_post"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, authorOfPost)
	respWriter = httptest.NewRecorder()
	testHandler.NewComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return

	}

	//  коммент добавлен
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	post := &post.Post{
		Score: 1,
		Views: 2,
		Type:  "text",
		Title: "fef",
		Author: &user.User{
			ID:       "310ca263",
			Username: "hhhhhhhh",
		},
		Category: "programming",
		Text:     "rferfer",
		Votes: []*vote.Vote{
			{
				Value:  1,
				UserID: "310ca263",
			},
		},

		Comments: []*comment.Comment{
			{
				Created: "2023-11-11T17:16:36.381Z",
				Author: &user.User{
					ID:       "GgGHsZysctdVTaCvTZWhgzLReBThTXHc",
					Username: "jjjjjjjj",
				},
				Body: "vrfer",
				ID:   "SVukaIqtTrARYDjQEmbICqFgRQpTHHrZ",
			},
		},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}

	authorOfComment := &user.User{
		ID:       "GgGHsZysctdVTaCvTZWhgzLReBThTXHc",
		Username: "jjjjjjjj",
	}

	testRepo.EXPECT().AddComment("some comment", authorOfComment, "654f63e3a2414a2a554b6423").Return(post, nil)
	request = httptest.NewRequest(http.MethodPost, "/api/post/654f63e3a2414a2a554b6423",
		strings.NewReader(`{"comment":"some comment"}`))
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "654f63e3a2414a2a554b6423"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, authorOfComment)
	respWriter = httptest.NewRecorder()
	testHandler.NewComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 201 {
		t.Errorf("expected status %d, got status %d", http.StatusCreated, resp.StatusCode)
	}

	expectedResp := `{"score":1,"views":2,"type":"text","title":"fef","author":{"id":"310ca263","username":"hhhhhhhh"},"category":"programming","text":"rferfer","votes":[{"vote":1,"user":"310ca263"}],"comments":[{"created":"2023-11-11T17:16:36.381Z","author":{"id":"GgGHsZysctdVTaCvTZWhgzLReBThTXHc","username":"jjjjjjjj"},"body":"vrfer","id":"SVukaIqtTrARYDjQEmbICqFgRQpTHHrZ"}],"created":"2023-11-11T14:22:11.695Z","upvotePercentage":100,"id":"654f63e3a2414a2a554b6423"}`
	if string(body) != expectedResp {
		t.Errorf("wrong response body,\nexpected: %s\n, got %s", expectedResp, string(body))
		return

	}

}

func TestPostHandlerDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	//  в контексте лежит битый юзер
	request := httptest.NewRequest(http.MethodDelete, "/api/post/dwdwdwe/rhbfhrb", nil)
	ctx := request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, "bad user value")
	respWriter := httptest.NewRecorder()
	testHandler.DeleteComment(respWriter, request.WithContext(ctx))
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return
	}

	currentUser := &user.User{
		ID:       "GgGHsZysctdVTaCvTZWhgzLReBThTXHc",
		Username: "jjjjjjjj",
	}

	//  удалить пытается юзер, не являющийся автором коммента
	testRepo.EXPECT().DeleteComment("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe", "hrgyfrfb").Return(nil, post.ErrNoAccess)
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe/hrgyfrfb", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe", "COMMENT_ID": "hrgyfrfb"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeleteComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 403 {
		t.Errorf("expected status %d, got status %d", http.StatusForbidden, resp.StatusCode)
		return

	}

	//  пост не найден
	testRepo.EXPECT().DeleteComment("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe", "hrgyfrfb").Return(nil, post.ErrNoPost)
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe/hrgyfrfb", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe", "COMMENT_ID": "hrgyfrfb"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeleteComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return

	}

	//  коммент не найден
	testRepo.EXPECT().DeleteComment("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe", "hrgyfrfb").Return(nil, post.ErrNoComment)
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe/hrgyfrfb", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe", "COMMENT_ID": "hrgyfrfb"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeleteComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return

	}

	//  какая то ошибка сервера
	testRepo.EXPECT().DeleteComment("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe", "hrgyfrfb").Return(nil, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe/hrgyfrfb", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe", "COMMENT_ID": "hrgyfrfb"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeleteComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return

	}

	// коммент успешно удален
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	postWithDeletedComment := &post.Post{
		Score: 1,
		Views: 0,
		Type:  "text",
		Title: "fef",
		Author: &user.User{
			ID:       "310ca263",
			Username: "hhhhhhhh",
		},
		Category: "programming",
		Text:     "rferfer",
		Votes: []*vote.Vote{
			{
				Value:  1,
				UserID: "310ca263",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}

	testRepo.EXPECT().DeleteComment("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "654f63e3a2414a2a554b6423", "hrgyfrfb").
		Return(postWithDeletedComment, nil)
	request = httptest.NewRequest(http.MethodDelete, "/api/post/654f63e3a2414a2a554b6423/hrgyfrfb", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "654f63e3a2414a2a554b6423", "COMMENT_ID": "hrgyfrfb"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeleteComment(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}

	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
		return

	}

}

func TestPostHandlerMakeVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	//  в контексте лежит битый юзер
	request := httptest.NewRequest(http.MethodGet, "/api/post/dwdwdw/upvote", nil)
	ctx := request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, "bad user value")
	respWriter := httptest.NewRecorder()
	testHandler.MakeVote(respWriter, request.WithContext(ctx))
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return
	}

	//  такого поста нет

	currentUser := &user.User{
		ID:       "GgGHsZysctdVTaCvTZWhgzLReBThTXHc",
		Username: "jjjjjjjj",
	}
	testRepo.EXPECT().UpVote("feygfyfe", "GgGHsZysctdVTaCvTZWhgzLReBThTXHc").Return(nil, post.ErrNoPost)
	request = httptest.NewRequest(http.MethodGet, "/api/post/feygfyfe/upvote", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.MakeVote(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return

	}

	//  какая то ошибка сервера
	testRepo.EXPECT().UnVote("feygfyfe", "GgGHsZysctdVTaCvTZWhgzLReBThTXHc").Return(nil, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodGet, "/api/post/feygfyfe/unvote", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.MakeVote(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return

	}

	//  оценка изменена
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	postWithDownVote := &post.Post{
		Score: -1,
		Views: 2,
		Type:  "text",
		Title: "fef",
		Author: &user.User{
			ID:       "310ca263",
			Username: "hhhhhhhh",
		},
		Category: "programming",
		Text:     "rferfer",
		Votes: []*vote.Vote{
			{
				Value:  -1,
				UserID: "GgGHsZysctdVTaCvTZWhgzLReBThTXHc",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 0,
		ID:               objID,
	}

	testRepo.EXPECT().DownVote("654f63e3a2414a2a554b6423", "GgGHsZysctdVTaCvTZWhgzLReBThTXHc").Return(postWithDownVote, nil)
	request = httptest.NewRequest(http.MethodGet, "/api/post/654f63e3a2414a2a554b6423/downvote", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "654f63e3a2414a2a554b6423"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.MakeVote(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
		return

	}

}

func TestPostHandlerDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}

	//  в контексте лежит битый юзер
	request := httptest.NewRequest(http.MethodGet, "/api/post/dwdwdw", nil)
	ctx := request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, "bad user value")
	respWriter := httptest.NewRecorder()
	testHandler.DeletePost(respWriter, request.WithContext(ctx))
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return
	}

	currentUser := &user.User{
		ID:       "GgGHsZysctdVTaCvTZWhgzLReBThTXHc",
		Username: "jjjjjjjj",
	}
	//  удалить пытается юзер, не являющийся автором поста
	testRepo.EXPECT().DeletePost("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe").Return(false, post.ErrNoAccess)
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeletePost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 403 {
		t.Errorf("expected status %d, got status %d", http.StatusForbidden, resp.StatusCode)
		return

	}

	//  пост не найден
	testRepo.EXPECT().DeletePost("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe").Return(false, post.ErrNoPost)
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeletePost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return

	}

	//  какая то ошибка сервера
	testRepo.EXPECT().DeletePost("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe").Return(false, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeletePost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return

	}

	//  пост удален
	testRepo.EXPECT().DeletePost("GgGHsZysctdVTaCvTZWhgzLReBThTXHc", "feygfyfe").Return(true, nil)
	request = httptest.NewRequest(http.MethodDelete, "/api/post/feygfyfe", nil)
	request = mux.SetURLVars(request, map[string]string{"POST_ID": "feygfyfe"})
	ctx = request.Context()
	ctx = context.WithValue(ctx, middleware.MyUserKey, currentUser)
	respWriter = httptest.NewRecorder()
	testHandler.DeletePost(respWriter, request.WithContext(ctx))
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
		return

	}

}

func TestPostHandlerListByUserLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testRepo := post.NewMockPostRepo(ctrl)
	testHandler := &PostHandler{
		Logger:   zap.NewNop().Sugar(),
		PostRepo: testRepo,
	}
	//  юзер не найден

	testRepo.EXPECT().GetPostsByUserID("username_not_exist").Return(nil, user.ErrNoUser)
	request := httptest.NewRequest(http.MethodGet, "/api/user/username_not_exist", nil)
	request = mux.SetURLVars(request, map[string]string{"USER_LOGIN": "username_not_exist"})
	respWriter := httptest.NewRecorder()
	testHandler.ListByUserLogin(respWriter, request)
	resp := respWriter.Result()
	_, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status %d, got status %d", http.StatusNotFound, resp.StatusCode)
		return

	}
	//  какая то ошибка сервера
	testRepo.EXPECT().GetPostsByUserID("username").Return(nil, fmt.Errorf("error"))
	request = httptest.NewRequest(http.MethodGet, "/api/user/username", nil)
	request = mux.SetURLVars(request, map[string]string{"USER_LOGIN": "username"})
	respWriter = httptest.NewRecorder()
	testHandler.ListByUserLogin(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 500 {
		t.Errorf("expected status %d, got status %d", http.StatusInternalServerError, resp.StatusCode)
		return

	}

	//  посты юзера найдены
	testRepo.EXPECT().GetPostsByUserID("username").Return([]*post.Post{}, nil)
	request = httptest.NewRequest(http.MethodGet, "/api/user/username", nil)
	request = mux.SetURLVars(request, map[string]string{"USER_LOGIN": "username"})
	respWriter = httptest.NewRecorder()
	testHandler.ListByUserLogin(respWriter, request)
	resp = respWriter.Result()
	_, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unable to read response body")
		return
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status %d, got status %d", http.StatusOK, resp.StatusCode)
		return

	}
}
