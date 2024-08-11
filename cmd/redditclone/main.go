package main

import (
	"context"
	"database/sql"
	"github.com/gomodule/redigo/redis"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"html/template"
	"log"
	"net/http"
	"reddit/pkg/handlers"
	"reddit/pkg/idgenerator"
	"reddit/pkg/middleware"
	"reddit/pkg/post"
	"reddit/pkg/session"
	"reddit/pkg/user"
)

func openMysql() (*sql.DB, error) {
	dsn := "root:"
	// mysqlPassword := os.Getenv("PASSWORDMYSQL")
	mysqlPassword := "mysql111"
	dsn += mysqlPassword
	dsn += "@tcp(mysql:3306)/golang?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	err = db.Ping()
	if err == nil {
		return db, nil
	}
	return nil, err
}

func openMongoDB() (*mongo.Client, error) {
	ctx := context.Background()
	sess, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://mongodb"))
	if err != nil {
		return nil, err
	}
	return sess, nil
}

func openRedis() (redis.Conn, error) {
	c, err := redis.DialURL("redis://user:@redis:6379/0")
	if err != nil {
		return nil, err
	}
	return c, nil

}

func main() {
	myTemplate := template.Must(template.ParseGlob("./06_databases/99_hw/redditclone/static/html/*"))
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Printf("error in logger initialization: %s", err)
		return
	}
	defer func(zapLogger *zap.Logger) {
		err = zapLogger.Sync()
		if err != nil {
			log.Printf("error in logger close: %s", err)
			return
		}
	}(zapLogger)
	logger := zapLogger.Sugar()

	dbSQL, err := openMysql()
	if err != nil {
		logger.Infof("error on connection to mysql: %s", err.Error())
		return
	}
	mongoSession, err := openMongoDB()
	if err != nil {
		logger.Infof("error on connection to mongoDB: %s", err.Error())
	}
	redisConn, err := openRedis()
	if err != nil {
		logger.Infof("error on connection to redis: %s", err.Error())
	}
	defer func(redisConn redis.Conn) {
		err = redisConn.Close()
		if err != nil {
			logger.Infof("error on redis close: %s", err.Error())
		}
	}(redisConn)

	dbMongoCollection := mongoSession.Database("golang").Collection("items")
	sessManMysql := session.SessionManagerMysql{
		DB: dbSQL,
	}
	sessManRedis := session.SessionManagerRedis{
		RedisConn: redisConn,
	}
	sessManDB := session.SessionManagerDB{
		SessionManagerMS:  sessManMysql,
		SessionManagerRDS: sessManRedis,
	}
	sessionManager := session.NewSessionManager(sessManDB)

	collectionHelper := &post.MongoCollection{
		Coll: dbMongoCollection,
	}
	clientHelper := &post.MongoClient{
		Cl: mongoSession,
	}
	postDBRepo := post.PostDBRepo{
		Posts: collectionHelper,
		Sess:  clientHelper,
	}
	userDBRepo := user.UserDBRepo{
		DB: dbSQL,
	}
	IDGenerator := &idgenerator.RandomIDGenerator{}

	userRepo := user.NewUserMemoryRepository(&userDBRepo, IDGenerator)
	postRepo := post.NewPostBusinessLogic(&postDBRepo, IDGenerator)

	userHandler := handlers.UserHandler{
		UserRepo:       userRepo,
		SessionManager: sessionManager,
		Logger:         logger,
	}

	postHandler := handlers.PostHandler{
		PostRepo: postRepo,
		Logger:   logger,
	}

	router := mux.NewRouter()

	staticRouter := router.PathPrefix("/static/").Subrouter()
	staticDir := "./06_databases/99_hw/redditclone/static"
	staticRouter.PathPrefix("/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	router.HandleFunc("/api/posts/", postHandler.List).Methods(http.MethodGet)
	router.HandleFunc("/api/posts/{CATEGORY_NAME}", postHandler.ListByCategory).Methods(http.MethodGet)
	router.HandleFunc("/api/post/{POST_ID}", postHandler.GetPostInfo).Methods(http.MethodGet)
	router.HandleFunc("/api/user/{USER_LOGIN}", postHandler.ListByUserLogin).Methods(http.MethodGet)

	router.HandleFunc("/api/register", userHandler.Register).Methods(http.MethodPost)
	router.HandleFunc("/api/login", userHandler.Login).Methods(http.MethodPost)

	// нужна авторизация

	rAuth := mux.NewRouter()
	router.Handle("/api/post/{POST_ID}/{COMMENT_ID}", middleware.Auth(logger, sessionManager, rAuth)).Methods(http.MethodDelete)
	router.Handle("/api/post/{POST_ID}/upvote", middleware.Auth(logger, sessionManager, rAuth)).Methods(http.MethodGet)
	router.Handle("/api/post/{POST_ID}/downvote", middleware.Auth(logger, sessionManager, rAuth)).Methods(http.MethodGet)
	router.Handle("/api/post/{POST_ID}/unvote", middleware.Auth(logger, sessionManager, rAuth)).Methods(http.MethodGet)
	router.Handle("/api/post/{POST_ID}", middleware.Auth(logger, sessionManager, rAuth)).Methods(http.MethodDelete)
	router.Handle("/api/posts", middleware.Auth(logger, sessionManager, rAuth)).Methods(http.MethodPost)
	router.Handle("/api/post/{POST_ID}", middleware.Auth(logger, sessionManager, rAuth)).Methods(http.MethodPost)

	rAuth.HandleFunc("/api/post/{POST_ID}/{COMMENT_ID}", postHandler.DeleteComment).Methods(http.MethodDelete)
	rAuth.HandleFunc("/api/post/{POST_ID}/upvote", postHandler.MakeVote).Methods(http.MethodGet)
	rAuth.HandleFunc("/api/post/{POST_ID}/downvote", postHandler.MakeVote).Methods(http.MethodGet)
	rAuth.HandleFunc("/api/post/{POST_ID}/unvote", postHandler.MakeVote).Methods(http.MethodGet)
	rAuth.HandleFunc("/api/post/{POST_ID}", postHandler.DeletePost).Methods(http.MethodDelete)
	rAuth.HandleFunc("/api/posts", postHandler.NewPost).Methods(http.MethodPost)
	rAuth.HandleFunc("/api/post/{POST_ID}", postHandler.NewComment).Methods(http.MethodPost)

	accessLogRouter := middleware.AccessLog(logger, router)
	errorLogRouter := middleware.ErrorLog(logger, accessLogRouter)
	mux := middleware.Panic(logger, errorLogRouter)

	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err = myTemplate.ExecuteTemplate(w, "index.html", struct{}{})
		if err != nil {
			http.Error(w, `Template errror`, http.StatusInternalServerError)
		}
	})

	addr := ":8042"

	logger.Infow("starting server",
		"type", "START",
		"addr", addr,
	)
	err = http.ListenAndServe(addr, mux)
	if err != nil {
		logger.Infof("errror in server start")
		return
	}
}
