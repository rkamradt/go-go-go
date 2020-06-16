package main

import (
	  "os"
		"context"
		"log"
		"time"
	  "github.com/gin-gonic/gin"
		"go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type Insert struct {
	Status string
	TotalResults int
	Articles [] Article
}
type Article struct {
		Source Source
		Author string
		Title string
		Description string
		Url string
		UrlToImage string
		PublishedAt string
		Content string
}
type Source struct {
		Id string
		Name string
}


func main() {
	mongoUser := os.Getenv("MONGO_USER")
	mongoPassword := os.Getenv("MONGO_PASS")

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://" +
					mongoUser +
					":" +
					mongoPassword +
					"@mongodb:27017"))
	if err != nil {
	    log.Fatal(err)
	}
	ctx, _ := context.WithTimeout(
		context.Background(),
		10*time.Second)
	err = client.Connect(ctx)

	collection := client.Database("news").Collection("inserts")

	r := gin.Default()
	r.GET("/v1/headlines", func(c *gin.Context) {
		ctx, _ = context.WithTimeout(
			context.Background(),
			5*time.Second)
		// Pass these options to the Find method
		findOptions := options.Find()
		findOptions.SetLimit(1000)

		// Here's an array in which you can store the decoded documents
		var results []*Article

		// Passing bson.D{{}} as the filter matches all documents in the collection
		cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
		if err != nil {
		    log.Fatal(err)
		}
		fromString := c.DefaultQuery("from", "")
		toString := c.DefaultQuery("to", "")
		// Finding multiple documents returns a cursor
		// Iterating through the cursor allows us to decode documents one at a time
		for cur.Next(context.TODO()) {

		    // create a value into which the single document can be decoded
		    var elem Insert
		    err := cur.Decode(&elem)
		    if err != nil {
		        log.Printf(err.Error())
		    }
        for _, article := range elem.Articles {
					if filterByDate(article, fromString, toString) {
						results = append(results, &article)
					}
				}
		}

		if err := cur.Err(); err != nil {
		    log.Printf(err.Error())
		}

		// Close the cursor once finished
		cur.Close(context.TODO())

		c.JSON(200, results)
	})
	r.Run() // listen and serve localhost:8080
}

func filterByDate(record Article, fromString string, toString string) bool {
		from, err := time.Parse(time.RFC3339, fromString)
		if err != nil {
			log.Printf("unable to parse from time " + fromString)
			return false
		}
		to, err := time.Parse(time.RFC3339, toString)
		if err != nil {
			log.Printf("unable to parse to time " + toString)
			return false
		}

		theDate, err := time.Parse(time.RFC3339, record.PublishedAt)
		if err != nil {
			log.Printf("unable to parse time " + record.PublishedAt)
			return false;
		}
		return theDate.After(from) &&
				theDate.Before(to);
}
