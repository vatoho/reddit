package post

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"reddit/pkg/comment"
	"reddit/pkg/idgenerator"
	"reddit/pkg/user"
	"reddit/pkg/vote"
)

func TestGetAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// какая то ошибка в монго
	testCollection.EXPECT().Find(context.Background(), bson.M{}).Return(nil, fmt.Errorf("error"))
	_, err := testRepo.GetAll()
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	postToReturn := &Post{
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
	posts := []interface{}{
		postToReturn,
	}

	// успешный запрос
	cursor, err := mongo.NewCursorFromDocuments(posts, nil, nil)
	if err != nil {
		t.Fatalf("error on cursor creation")
		return
	}
	testCollection.EXPECT().Find(context.Background(), bson.M{}).Return(cursor, nil)
	_, err = testRepo.GetAll()
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

}

func TestAddPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// ошибка добавления поста в бд
	testCollection.EXPECT().InsertOne(context.Background(), gomock.Any()).Return(nil, fmt.Errorf("error"))
	postToAdd := &Post{
		Type:     "text",
		Title:    "fef",
		Category: "programming",
		Text:     "rferfer",
	}
	author := &user.User{
		ID:       "310ca263",
		Username: "hhhhhhhh",
	}
	_, err := testRepo.AddPost(postToAdd, author)
	if err == nil {
		t.Errorf("ecpected error, got nil")
		return
	}

	// пост добавлен
	testCollection.EXPECT().InsertOne(context.Background(), gomock.Any()).Return("any", nil)
	postToAdd = &Post{
		Type:     "text",
		Title:    "fef",
		Category: "programming",
		Text:     "rferfer",
	}
	author = &user.User{
		ID:       "310ca263",
		Username: "hhhhhhhh",
	}
	postToAdd.Type = "link"

	_, err = testRepo.AddPost(postToAdd, author)
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

}

func TestGetPostByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// какая то ошибка в монго
	testCollection.EXPECT().Find(context.Background(), bson.M{"category": "programming"}).Return(nil, fmt.Errorf("error"))
	_, err := testRepo.GetPostByCategory("programming")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	postToReturn := &Post{
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
	posts := []interface{}{
		postToReturn,
	}

	// успешный запрос
	cursor, err := mongo.NewCursorFromDocuments(posts, nil, nil)
	if err != nil {
		t.Fatalf("error on cursor creation")
		return
	}
	testCollection.EXPECT().Find(context.Background(), bson.M{"category": "programming"}).Return(cursor, nil)
	_, err = testRepo.GetPostByCategory("programming")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

}

func TestGetPostByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// какая то ошибка в монго

	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	singleResponse := mongo.NewSingleResultFromDocument(nil, fmt.Errorf("error"), nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.GetPostByID("654f63e3a2414a2a554b6423")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// некорректный айди

	_, err = testRepo.GetPostByID("некорректный айди")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if !errors.Is(err, ErrNoPost) {
		t.Errorf("wrong error: expected %s, got %s", ErrNoPost, err)
		return
	}

	postToReturn := &Post{
		Score: 1,
		Views: 1,
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

	// пост найден но не получается увеличить счетчик просмотров

	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": bson.M{"views": 2}}).Return(nil, fmt.Errorf("db_error"))
	_, err = testRepo.GetPostByID("654f63e3a2414a2a554b6423")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// пот найден, просмотры обновлены
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), bson.M{"_id": objID}, bson.M{"$set": bson.M{"views": 2}}).Return(nil, nil)
	postWithUpdatedViews, err := testRepo.GetPostByID("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if postWithUpdatedViews.Views != postToReturn.Views+1 {
		t.Errorf("bad value of views: expected %d, got %d", postToReturn.Views+1, postWithUpdatedViews.Views)
		return
	}

}

func TestAddComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// не получилось найти пост
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	singleResponse := mongo.NewSingleResultFromDocument(nil, mongo.ErrNilDocument, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.AddComment("comment", &user.User{}, "654f63e3a2414a2a554b6423")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// пост найден, но не удалось добавить коммент
	postToReturn := &Post{
		Score: 1,
		Views: 1,
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

	authorOfComment := &user.User{
		ID:       "user_id",
		Username: "username",
	}

	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	filter := bson.M{"_id": objID}
	testCollection.EXPECT().UpdateOne(context.TODO(), filter, gomock.Any()).Return(nil, fmt.Errorf("db_error"))
	_, err = testRepo.AddComment("new_comment", authorOfComment, "654f63e3a2414a2a554b6423")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// коммент успешно добавлен

	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.TODO(), filter, gomock.Any()).Return(nil, nil)
	postWithNewComment, err := testRepo.AddComment("new_comment", authorOfComment, "654f63e3a2414a2a554b6423")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(postWithNewComment.Comments) == 0 {
		t.Errorf("wrong number of comments: expected 1, got %d", len(postWithNewComment.Comments))
		return
	}

}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// не получилось найти пост
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	singleResponse := mongo.NewSingleResultFromDocument(nil, mongo.ErrNilDocument, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.DeleteComment("user_id", "654f63e3a2414a2a554b6423", "comment_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// у поста нет комментария с таким айди
	commentAuthor := &user.User{
		ID:       "user_id",
		Username: "username",
	}
	postToReturn := &Post{
		Score: 1,
		Views: 1,
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
				ID:     "comment_id",
				Author: commentAuthor,
			},
		},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.DeleteComment("user_id", "654f63e3a2414a2a554b6423", "comment_id_wrong")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if !errors.Is(err, ErrNoComment) {
		t.Errorf("expected error %s, got %s", ErrNoComment, err)
		return
	}

	// у поста есть такой комментарий, но его хочет удалить не его автор
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.DeleteComment("user_id_wrong", "654f63e3a2414a2a554b6423", "comment_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	if !errors.Is(err, ErrNoAccess) {
		t.Errorf("expected error %s, got %s", ErrNoAccess, err)
		return
	}

	// не получается удалить коммент из базы данных
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.TODO(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("dr_error"))
	_, err = testRepo.DeleteComment("user_id", "654f63e3a2414a2a554b6423", "comment_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// коммент успешно удален
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.TODO(), gomock.Any(), gomock.Any()).Return(nil, nil)
	post, err := testRepo.DeleteComment("user_id", "654f63e3a2414a2a554b6423", "comment_id")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(post.Comments) != len(postToReturn.Comments)-1 {
		t.Errorf("wrong number of comments: expected %d, got %d", len(postToReturn.Comments)-1, len(post.Comments))
		return
	}

}

func TestUpvote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// не получилось найти пост
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	singleResponse := mongo.NewSingleResultFromDocument(nil, mongo.ErrNilDocument, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.UpVote("654f63e3a2414a2a554b6423", "user_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на этом посте уже стоит upvote
	postToReturn := &Post{
		Score: 1,
		Views: 1,
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
				UserID: "user_id",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	post, err := testRepo.UpVote("654f63e3a2414a2a554b6423", "user_id")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if !reflect.DeepEqual(post, postToReturn) {
		t.Errorf("wrong post: expected %v, got %v", postToReturn, post)
		return
	}

	// на посте стоит downvote, но не получается в базу записать upvote
	postToReturn = &Post{
		Score: 1,
		Views: 1,
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
				UserID: "user_id",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
	_, err = testRepo.UpVote("654f63e3a2414a2a554b6423", "user_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на посте нет оценок от этого пользователя, не получается добавить upvote
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
	_, err = testRepo.UpVote("654f63e3a2414a2a554b6423", "user_id_another")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на посте нет оценок пользователя, добавление успешно
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, nil)
	post, err = testRepo.UpVote("654f63e3a2414a2a554b6423", "user_id_another")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(post.Votes) != len(postToReturn.Votes)+1 {
		t.Errorf("wrong number of votes: expected %d, got %d", len(postToReturn.Votes)+1, len(post.Votes))
		return
	}

}

func TestDownVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// не получилось найти пост
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	singleResponse := mongo.NewSingleResultFromDocument(nil, mongo.ErrNilDocument, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.DownVote("654f63e3a2414a2a554b6423", "user_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на этом посте уже стоит downvote
	postToReturn := &Post{
		Score: 1,
		Views: 1,
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
				UserID: "user_id",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	post, err := testRepo.DownVote("654f63e3a2414a2a554b6423", "user_id")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if !reflect.DeepEqual(post, postToReturn) {
		t.Errorf("wrong post: expected %v, got %v", postToReturn, post)
		return
	}

	// на посте стоит upvote, но не получается в базу записать downvote
	postToReturn = &Post{
		Score: 1,
		Views: 1,
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
				UserID: "user_id",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
	_, err = testRepo.DownVote("654f63e3a2414a2a554b6423", "user_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на посте нет оценок от этого пользователя, не получается добавить downvote
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
	_, err = testRepo.DownVote("654f63e3a2414a2a554b6423", "user_id_another")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на посте нет оценок пользователя, добавление успешно
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, nil)
	post, err = testRepo.DownVote("654f63e3a2414a2a554b6423", "user_id_another")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if len(post.Votes) != len(postToReturn.Votes)+1 {
		t.Errorf("wrong number of votes: expected %d, got %d", len(postToReturn.Votes)+1, len(post.Votes))
		return
	}

}

func TestUnVote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// не получилось найти пост
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	singleResponse := mongo.NewSingleResultFromDocument(nil, mongo.ErrNilDocument, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.UnVote("654f63e3a2414a2a554b6423", "user_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на этом посте и так нет оценок от этого юзера
	postToReturn := &Post{
		Score: 1,
		Views: 1,
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
				UserID: "user_id",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	post, err := testRepo.UnVote("654f63e3a2414a2a554b6423", "user_id_another")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}
	if !reflect.DeepEqual(post, postToReturn) {
		t.Errorf("wrong post: expected %v, got %v", postToReturn, post)
		return
	}

	// на посте есть оценка этого юзера, но в базу не получается отправить изменения
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("error"))
	_, err = testRepo.UnVote("654f63e3a2414a2a554b6423", "user_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// на посте есть upvote, после ужаления на посте еще есть оценки

	postToReturn.Votes = append(postToReturn.Votes, &vote.Vote{
		Value: 1,
	})
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, nil)
	_, err = testRepo.UnVote("654f63e3a2414a2a554b6423", "user_id")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

	// на посте есть анвоут
	postToReturn = &Post{
		Score: 1,
		Views: 1,
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
				UserID: "user_id",
			},
			{
				Value:  1,
				UserID: "some_id",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().UpdateOne(context.Background(), gomock.Any(), gomock.Any()).Return(nil, nil)
	_, err = testRepo.UnVote("654f63e3a2414a2a554b6423", "user_id")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// не получилось найти пост
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	singleResponse := mongo.NewSingleResultFromDocument(nil, mongo.ErrNilDocument, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.DeletePost("user_id", "654f63e3a2414a2a554b6423")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// тот кто хочет удалить пост не является его автором
	postToReturn := &Post{
		Score: 1,
		Views: 1,
		Type:  "text",
		Title: "fef",
		Author: &user.User{
			ID:       "user_id",
			Username: "hhhhhhhh",
		},
		Category: "programming",
		Text:     "rferfer",
		Votes: []*vote.Vote{
			{
				Value:  -1,
				UserID: "user_id",
			},
			{
				Value:  1,
				UserID: "some_id",
			},
		},

		Comments:         []*comment.Comment{},
		Created:          "2023-11-11T14:22:11.695Z",
		UpvotePercentage: 100,
		ID:               objID,
	}
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	_, err = testRepo.DeletePost("user_id_another", "654f63e3a2414a2a554b6423")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if !errors.Is(err, ErrNoAccess) {
		t.Errorf("wrong error: expected %s, got %s", ErrNoAccess, err)
		return
	}

	// ошибка удаления поста
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().DeleteOne(context.Background(), gomock.Any()).Return(int64(0), fmt.Errorf("error"))
	_, err = testRepo.DeletePost("user_id", "654f63e3a2414a2a554b6423")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// пост успешно удален
	singleResponse = mongo.NewSingleResultFromDocument(postToReturn, nil, nil)
	testCollection.EXPECT().FindOne(context.Background(), bson.M{"_id": objID}).Return(singleResponse)
	testCollection.EXPECT().DeleteOne(context.Background(), gomock.Any()).Return(int64(1), nil)
	_, err = testRepo.DeletePost("user_id", "654f63e3a2414a2a554b6423")
	if err != nil {
		t.Errorf("enexpected error: %s", err)
		return
	}

}

func TestGetPostByUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	testCollection := NewMockCollectionHelper(ctrl)
	testRepoDB := &PostDBRepo{
		Posts: testCollection,
	}
	testIDGen := &idgenerator.TestIDGenerator{}
	testRepo := NewPostBusinessLogic(testRepoDB, testIDGen)

	// нет поста с заданным автором

	testCollection.EXPECT().Find(context.Background(), gomock.Any()).Return(nil, mongo.ErrNoDocuments)
	_, err := testRepo.GetPostsByUserID("user_id")
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}

	// есть посты этого юзера
	objID, err := primitive.ObjectIDFromHex("654f63e3a2414a2a554b6423")
	if err != nil {
		t.Fatalf("error in id")
		return
	}
	postToReturn := &Post{
		Score: 1,
		Views: 0,
		Type:  "text",
		Title: "fef",
		Author: &user.User{
			ID:       "user_id",
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
	posts := []interface{}{
		postToReturn,
	}
	cursor, err := mongo.NewCursorFromDocuments(posts, nil, nil)
	if err != nil {
		t.Fatalf("error in cursor creation")
		return
	}
	testCollection.EXPECT().Find(context.Background(), gomock.Any()).Return(cursor, nil)
	_, err = testRepo.GetPostsByUserID("user_id")
	if err != nil {
		t.Errorf("unexpected error: %s", err)
		return
	}

}
