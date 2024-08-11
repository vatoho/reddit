package post

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type DatabaseHelper interface {
	Collection(name string) CollectionHelper
	Client() ClientHelper
}

type CollectionHelper interface {
	Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error)
	FindOne(context.Context, interface{}) SingleResultHelper
	InsertOne(context.Context, interface{}) (interface{}, error)
	DeleteOne(ctx context.Context, filter interface{}) (int64, error)
	UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error)
}

type SingleResultHelper interface {
	Decode(v interface{}) error
}

type ClientHelper interface {
	Database(string) DatabaseHelper
	Connect() error
	StartSession() (mongo.Session, error)
}

type PostDBRepo struct {
	Posts CollectionHelper
	Sess  ClientHelper
}

func (p *PostDBRepo) IncreasePostViewsDB(post *Post, postID string) error {
	postIDMongo, err := getMongoID(postID)
	if err != nil {
		return err
	}
	_, err = p.Posts.UpdateOne(context.Background(), bson.M{"_id": postIDMongo}, bson.M{"$set": bson.M{"views": post.Views}})
	return err
}

func (p *PostDBRepo) GetPostByCategoryDB(postOfCurrentCategory []*Post, category string) ([]*Post, error) {
	result, err := p.Posts.Find(context.Background(), bson.M{"category": category})
	if err != nil {
		return nil, err
	}
	err = result.All(context.Background(), &postOfCurrentCategory)
	if err != nil {
		return nil, err
	}
	return postOfCurrentCategory, nil
}

func (p *PostDBRepo) AddPostDB(post *Post) error {
	post.ID = primitive.NewObjectID()
	_, err := p.Posts.InsertOne(context.Background(), post)
	return err
}

func (p *PostDBRepo) GetAllPostsDB(allPosts []*Post) ([]*Post, error) {
	result, err := p.Posts.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	err = result.All(context.Background(), &allPosts)
	if err != nil {
		return nil, err
	}
	return allPosts, nil
}

func (p *PostDBRepo) AddCommentDB(post *Post, postID string) error {
	postIDMongo, err := getMongoID(postID)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{"comments": post.Comments},
	}
	filter := bson.M{"_id": postIDMongo}
	_, err = p.Posts.UpdateOne(context.TODO(), filter, update)
	return err
}

func (p *PostDBRepo) DeleteCommentDB(postWithCommentToDelete *Post, postID string) error {
	postIDMongo, err := getMongoID(postID)
	if err != nil {
		return err
	}
	update := bson.M{
		"$set": bson.M{"comments": postWithCommentToDelete.Comments},
	}
	filter := bson.M{"_id": postIDMongo}
	_, err = p.Posts.UpdateOne(context.TODO(), filter, update)
	return err
}

func (p *PostDBRepo) GetPostByIDDB(postID string) (*Post, error) {
	postIDMongo, err := getMongoID(postID)
	if err != nil {
		return nil, err
	}
	post := &Post{}
	err = p.Posts.FindOne(context.Background(), bson.M{"_id": postIDMongo}).Decode(post)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (p *PostDBRepo) SetPostDB(postToSet *Post, postID string) error {
	postIDMongo, err := getMongoID(postID)
	if err != nil {
		return err
	}
	filter := bson.M{"_id": postIDMongo}
	update := bson.M{
		"$set": postToSet,
	}
	_, err = p.Posts.UpdateOne(context.Background(), filter, update)
	return err
}

func (p *PostDBRepo) GetPostByUsernameDB(userName string) ([]*Post, error) {
	userPosts := make([]*Post, 0)
	result, err := p.Posts.Find(context.Background(), bson.M{"author.username": userName})
	if err != nil {
		return nil, err
	}
	err = result.All(context.Background(), &userPosts)
	if err != nil {
		return nil, err
	}
	return userPosts, nil
}

func (p *PostDBRepo) DeletePostDB(postID string) (bool, error) {
	postIDMongo, err := getMongoID(postID)
	if err != nil {
		return false, err
	}
	_, err = p.Posts.DeleteOne(context.Background(), bson.M{"_id": postIDMongo})
	if err != nil {
		return false, err
	}
	return true, nil
}

func getMongoID(id string) (primitive.ObjectID, error) {
	postIDMongo, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return primitive.NilObjectID, ErrNoPost
	}
	return postIDMongo, nil
}

type MongoClient struct {
	Cl *mongo.Client
}
type MongoDatabase struct {
	DB *mongo.Database
}
type MongoCollection struct {
	Coll *mongo.Collection
}

type MongoSingleResult struct {
	Sr *mongo.SingleResult
}

type MongoSession struct {
	mongo.Session
}

func (mc *MongoClient) Database(dbName string) DatabaseHelper {
	db := mc.Cl.Database(dbName)
	return &MongoDatabase{DB: db}
}

func (mc *MongoClient) StartSession() (mongo.Session, error) {
	session, err := mc.Cl.StartSession()
	return &MongoSession{session}, err
}

func (mc *MongoClient) Connect() error {
	return mc.Cl.Connect(nil)
}

func (md *MongoDatabase) Collection(colName string) CollectionHelper {
	collection := md.DB.Collection(colName)
	return &MongoCollection{Coll: collection}
}

func (md *MongoDatabase) Client() ClientHelper {
	client := md.DB.Client()
	return &MongoClient{Cl: client}
}

func (mc *MongoCollection) Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error) {
	return mc.Coll.Find(ctx, filter)
}

func (mc *MongoCollection) FindOne(ctx context.Context, filter interface{}) SingleResultHelper {
	singleResult := mc.Coll.FindOne(ctx, filter)
	return &MongoSingleResult{Sr: singleResult}
}

func (mc *MongoCollection) InsertOne(ctx context.Context, document interface{}) (interface{}, error) {
	id, err := mc.Coll.InsertOne(ctx, document)
	return id.InsertedID, err
}

func (mc *MongoCollection) DeleteOne(ctx context.Context, filter interface{}) (int64, error) {
	count, err := mc.Coll.DeleteOne(ctx, filter)
	return count.DeletedCount, err
}

func (mc *MongoCollection) UpdateOne(ctx context.Context, filter interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return mc.Coll.UpdateOne(ctx, filter, update)
}

func (sr *MongoSingleResult) Decode(v interface{}) error {
	return sr.Sr.Decode(v)
}
