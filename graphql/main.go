package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"regexp"

	"cloud.google.com/go/storage"
	"github.com/graphql-go/graphql"
	"github.com/joho/godotenv"
	"github.com/rs/cors"

	// medium "github.com/medium/medium-sdk-go"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/olivere/elastic/v7"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

var jwtSecret []byte

var tokenExpiration int

var sendgridAPIKey string

var websiteURL string

var apiURL string

var mongoClient *mongo.Client

var ctxMongo context.Context

var mainDatabase = "website"

var userCollection *mongo.Collection

var userMongoName = "users"

var blogCollection *mongo.Collection

var blogMongoName = "blogs"

var projectCollection *mongo.Collection

var projectMongoName = "projects"

var shortLinkCollection *mongo.Collection

var shortLinkMongoName = "shortlink"

var elasticClient *elastic.Client

var ctxElastic context.Context

var blogElasticIndex = "blogs"

var blogElasticType = "blog"

var projectElasticIndex = "projects"

var projectElasticType = "project"

var blogType = "blog"

var projectType = "project"

var validTypes = []string{
	blogType,
	projectType,
}

var ctxStorage context.Context

var storageClient *storage.Client

var storageBucket *storage.BucketHandle

var blogFileIndex = "blogfiles"

var projectFileIndex = "projectfiles"

var placeholderPath = "/placeholder"

var originalPath = "/original"

var blurPath = "/blur"

var haveblur = []string{
	"image/jpeg",
	"image/png",
	"image/gif",
}

var validContentTypes = []string{
	"image/jpeg",
	"image/png",
	"image/svg+xml",
	"image/gif",
	"video/mpeg",
	"video/webm",
	"video/mp4",
	"video/x-msvideo",
	"application/pdf",
	"text/plain",
	"application/zip",
	"text/csv",
	"application/json",
	"application/ld+json",
	"application/vnd.ms-powerpoint",
	"application/vnd.openxmlformats-officedocument.presentationml.presentation",
	"application/msword",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
}

var progressiveImageSize = 30

var progressiveImageBlurAmount = 20.0

var logger *zap.Logger

type tokenKeyType string

var tokenKey tokenKeyType

var redisClient *redis.Client

var cacheTime time.Duration

var validHexcode *regexp.Regexp

var postSearchFields = []string{
	"title",
	"author",
	"caption",
	"content",
}

var mainRecaptchaSecret string

var shortlinkRecaptchaSecret string

var shortlinkURL string

var serviceEmail string

var jwtIssuer string

var lastSitemapUpdate string

var sitemapTimeFormat = "2006-01-02T15:04:05Z07:00"

var mode string

// var mediumClient *medium.Medium

// var mediumUser *medium.User

/**
 * @api {get} /hello Test rest request
 * @apiVersion 0.0.1
 * @apiSuccess {String} message Hello message
 * @apiGroup misc
 */
func hello(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	response.Write([]byte(`{"message":"Hello!"}`))
}

func main() {
	// "./logs"
	loggerconfig := []byte(`{
		"level": "debug",
		"encoding": "json",
		"outputPaths": ["stdout"],
		"errorOutputPaths": ["stderr"],
		"initialFields": {},
		"encoderConfig": {
		  "messageKey": "message",
		  "levelKey": "level",
		  "levelEncoder": "lowercase"
		}
  }`)
	var zapconfig zap.Config
	if err := json.Unmarshal(loggerconfig, &zapconfig); err != nil {
		panic(err)
	}
	var err error
	logger, err = zapconfig.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	logger.Info("logger created")
	err = godotenv.Load()
	if err != nil {
		logger.Fatal("Error loading .env file")
	}
	jwtSecret = []byte(os.Getenv("SECRET"))
	tokenExpiration, err = strconv.Atoi(os.Getenv("TOKENEXPIRATION"))
	if err != nil {
		logger.Fatal(err.Error())
	}
	sendgridAPIKey = os.Getenv("SENDGRIDAPIKEY")
	serviceEmail = os.Getenv("SERVICEEMAIL")
	jwtIssuer = os.Getenv("JWTISSUER")
	lastSitemapUpdateTime, err := time.Parse(sitemapTimeFormat, os.Getenv("LASTSITEMAPUPDATE"))
	if err != nil {
		logger.Fatal(err.Error())
	}
	lastSitemapUpdate = lastSitemapUpdateTime.Format(sitemapTimeFormat)
	mode = os.Getenv("MODE")
	websiteURL = os.Getenv("WEBSITEURL")
	apiURL = os.Getenv("APIURL")
	ctxMongo, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	cancel()
	mongouri := os.Getenv("MONGOURI")
	mongoClient, err = mongo.Connect(ctxMongo, options.Client().ApplyURI(mongouri))
	if err != nil {
		logger.Fatal(err.Error())
	}
	userCollection = mongoClient.Database(mainDatabase).Collection(userMongoName)
	projectCollection = mongoClient.Database(mainDatabase).Collection(projectMongoName)
	blogCollection = mongoClient.Database(mainDatabase).Collection(blogMongoName)
	shortLinkCollection = mongoClient.Database(mainDatabase).Collection(shortLinkMongoName)
	elasticuri := os.Getenv("ELASTICURI")
	elasticClient, err = elastic.NewClient(elastic.SetSniff(false), elastic.SetURL(elasticuri))
	if err != nil {
		logger.Fatal(err.Error())
	}
	ctxElastic = context.Background()
	var storageconfigstr = os.Getenv("STORAGECONFIG")
	var storageconfigjson map[string]interface{}
	json.Unmarshal([]byte(storageconfigstr), &storageconfigjson)
	ctxStorage = context.Background()
	storageconfigjsonbytes, err := json.Marshal(storageconfigjson)
	if err != nil {
		logger.Fatal(err.Error())
	}
	storageClient, err = storage.NewClient(ctxStorage, option.WithCredentialsJSON([]byte(storageconfigjsonbytes)))
	if err != nil {
		logger.Fatal(err.Error())
	}
	bucketName := os.Getenv("STORAGEBUCKETNAME")
	storageBucket = storageClient.Bucket(bucketName)
	gcpprojectid, ok := storageconfigjson["project_id"].(string)
	if !ok {
		logger.Fatal("could not cast gcp project id to string")
	}
	if err := storageBucket.Create(ctxStorage, gcpprojectid, nil); err != nil {
		logger.Info(err.Error())
	}
	redisAddress := os.Getenv("REDISADDRESS")
	redisPassword := os.Getenv("REDISPASSWORD")
	cacheSeconds, err := strconv.Atoi(os.Getenv("CACHETIME"))
	if err != nil {
		logger.Fatal(err.Error())
	}
	cacheTime = time.Duration(cacheSeconds) * time.Second
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddress,
		Password: redisPassword,
		DB:       0, // use default DB
	})
	pong, err := redisClient.Ping().Result()
	if err != nil {
		logger.Fatal(err.Error())
	} else {
		logger.Info("connected to redis cache: " + pong)
	}
	validHexcode, err = regexp.Compile("(^#[0-9A-F]{6}$)|(^#[0-9A-F]{8}$)")
	if err != nil {
		logger.Fatal(err.Error())
	}
	mainRecaptchaSecret = os.Getenv("MAINRECAPTCHASECRET")
	shortlinkRecaptchaSecret = os.Getenv("SHORTLINKRECAPTCHASECRET")
	shortlinkURL = os.Getenv("SHORTLINKURL")
	/*
		mediumAccessToken := os.Getenv("MEDIUMACCESSTOKEN")
		mediumClient = medium.NewClientWithAccessToken(mediumAccessToken)
		mediumUser, err := mediumClient.GetUser("")
		if err != nil {
			logger.Fatal(err.Error())
		}
		logger.Info("medium id " + mediumUser.ID)
	*/
	port := ":" + os.Getenv("PORT")
	schema, err := graphql.NewSchema(graphql.SchemaConfig{
		Query:    rootQuery(),
		Mutation: rootMutation(),
	})
	if err != nil {
		logger.Fatal(err.Error())
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/graphql", func(response http.ResponseWriter, request *http.Request) {
		tokenKey = tokenKeyType("token")
		var query = ""
		queryString := request.URL.Query().Get("query")
		if queryString != "" {
			query = queryString
		} else if request.Method == http.MethodPost || request.Method == http.MethodPut {
			logger.Info("got put or post")
			var querydata map[string]interface{}
			err = nil
			querybody, err := ioutil.ReadAll(request.Body)
			if err == nil {
				err = json.Unmarshal(querybody, &querydata)
				if err == nil {
					queryfromjson, ok := querydata["query"].(string)
					if ok {
						query = queryfromjson
					}
				}
			}
		}
		result := graphql.Do(graphql.Params{
			Schema:        schema,
			RequestString: query,
			Context:       context.WithValue(context.Background(), tokenKey, getAuthToken(request)),
		})
		response.Header().Set("Content-Type", "application/json")
		json.NewEncoder(response).Encode(result)
	})
	mux.HandleFunc("/countPosts", countPosts)
	mux.HandleFunc("/sendTestEmail", sendTestEmail)
	mux.HandleFunc("/loginEmailPassword", loginEmailPassword)
	mux.HandleFunc("/logoutEmailPassword", logoutEmailPassword)
	mux.HandleFunc("/register", register)
	mux.HandleFunc("/verify", verifyEmail)
	mux.HandleFunc("/sendResetEmail", sendPasswordResetEmail)
	mux.HandleFunc("/reset", resetPassword)
	mux.HandleFunc("/hello", hello)
	mux.HandleFunc("/getPostFile", getPostFile)
	mux.HandleFunc("/writePostFile", writePostFile)
	mux.HandleFunc("/deletePostFiles", deletePostFiles)
	mux.HandleFunc("/shortlink", shortLinkRedirect)
	mux.HandleFunc("/createShortLink", createShortLink)
	mux.HandleFunc("/sitemap.xml", sitemapIndex)
	mux.HandleFunc("/sitemap.xml.gz", sitemapIndexGZip)
	mux.HandleFunc("/sitemap-blogs.xml", sitemapBlogs)
	mux.HandleFunc("/sitemap-blogs.xml.gz", sitemapBlogsGZip)
	mux.HandleFunc("/sitemap-projects.xml", sitemapProjects)
	mux.HandleFunc("/sitemap-projects.xml.gz", sitemapProjectsGZip)
	var allowedOrigins []string
	if mode == "debug" {
		allowedOrigins = []string{
			"*",
		}
	} else {
		allowedOrigins = []string{
			websiteURL,
			shortlinkURL,
		}
	}
	thecors := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
			"OPTIONS",
		},
		OptionsPassthrough: false,
		Debug:              mode == "debug",
	})
	handler := thecors.Handler(mux)
	http.ListenAndServe(port, handler)
	logger.Info("Starting the application at " + port + " 🚀")
}

func getAuthToken(request *http.Request) string {
	authToken := request.Header.Get("Authorization")
	splitToken := strings.Split(authToken, "Bearer ")
	if splitToken != nil && len(splitToken) > 1 {
		authToken = splitToken[1]
	}
	return authToken
}

func validType(thetype string) bool {
	for _, b := range validTypes {
		if b == thetype {
			return true
		}
	}
	return false
}
