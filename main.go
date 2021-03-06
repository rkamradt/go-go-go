package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
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
  mongoHost := os.Getenv("MONGO_HOST")

	mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:27017",
    mongoUser,
    mongoPassword,
    mongoHost)
  client, err := mongo.NewClient(options.Client().ApplyURI(mongoURI))
  if err != nil {
      log.Fatal(err)
  }
  ctx, cancel := context.WithTimeout(
    context.Background(),
    10*time.Second)
  defer cancel()

  if err = client.Connect(ctx); err != nil {
      log.Fatal(err)
  }

  collection := client.Database("news").Collection("inserts")

  r := gin.Default()
  r.GET("/v1/headlines", func(c *gin.Context) {
    ctx, cancel = context.WithTimeout(
      context.Background(),
      5*time.Second)
    defer cancel()

    // Pass these options to the Find method
    findOptions := options.Find()
    findOptions.SetLimit(1000)

    // Here's an array in which you can store the decoded documents
    var results [] Article

    // Passing bson.D{{}} as the filter matches all documents in the collection
    cur, err := collection.Find(ctx, bson.D{{}}, findOptions)
    if err != nil {
        log.Print(err)
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    defer cur.Close(ctx)
    fromString := c.DefaultQuery("from", "")
    toString := c.DefaultQuery("to", "")
    for cur.Next(ctx) {

      var elem Insert
      err := cur.Decode(&elem)
      if err != nil {
          log.Printf(err.Error())
      } else {
        for _, article := range elem.Articles {
          if filterByDate(article, fromString, toString) {
            results = append(results, article)
          }
        }
      }
    }

    if err := cur.Err(); err != nil {
        log.Print(err)
    }
    c.JSON(200, results)
  })
  if err = r.Run(); err != nil {
    log.Panic(err)
  }
}

func filterByDate(record Article, fromString string, toString string) bool {
  from, err := time.Parse(time.RFC3339, fromString)
  noFrom := err != nil
  to, err := time.Parse(time.RFC3339, toString)
  noTo := err != nil

  theDate, err := time.Parse(time.RFC3339, record.PublishedAt)
  if err != nil {
    log.Printf("unable to parse time " + record.PublishedAt)
    return false;
  }
  return (noFrom || theDate.After(from)) &&
      (noTo || theDate.Before(to))
}
