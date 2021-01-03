package main

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"os"
	"time"
)

type Todo struct{
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Todo string `json:"todo" bson:"todo"`
	Description string `json:"description" bson:"description"`
	Status bool `json:"status" bson:"status"`

}


func main() {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders:  "Origin, Content-Type, Accept",
	}))
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error Loading .env file")
	}


	dbUsername :=os.Getenv("DB_USERNAME")
	dbPassword :=os.Getenv("DB_PASSWORD")
	dbName :=	os.Getenv("DB_NAME")


	url:="mongodb+srv://"+dbUsername+":"+dbPassword+"@cluster0.yejhx.mongodb.net/"+dbName+"?retryWrites=true&w=majority"




	app.Get("/", func(c *fiber.Ctx) error {
		var todos []Todo
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(url))
		if err != nil {
			log.Fatal("Error While connecting:  ",err)
		}
		collection:=client.Database("test").Collection("todos")
		cursor,err := collection.Find(ctx,bson.M{})
		if err != nil{
			log.Fatal(err)
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx){
			var todo Todo
			_ = cursor.Decode(&todo)
			todos = append(todos,todo)
		}
		defer client.Disconnect(ctx)
		return c.JSON(&todos)


	})

	app.Post("/",func (c *fiber.Ctx)error{
		var todo Todo
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client, _ := mongo.Connect(ctx, options.Client().ApplyURI(url))
		c.BodyParser(&todo)
		collection:= client.Database("test").Collection("todos")
		_, err := collection.InsertOne(ctx, todo)
		if err != nil{
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)
		return c.Status(201).SendString("Todo Added")

	})

	app.Put("/:id",func (c *fiber.Ctx)error{
		var todo Todo
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		client, _ := mongo.Connect(ctx, options.Client().ApplyURI(url))
		c.BodyParser(&todo)
		id,_ := primitive.ObjectIDFromHex(c.Params("id"))
		filter :=bson.M{"_id":	id}
		values := bson.D{{"$set",bson.D{{ "status", todo.Status}}}}
		collection:= client.Database("test").Collection("todos")
		_, err = collection.UpdateOne(ctx, filter,values)
		if err !=nil{
			log.Fatal("error while updating",err)
		}
		defer client.Disconnect(ctx)
		return c.Status(201).SendString("Todo Updated")

	})

	app.Delete("/:id",func (c *fiber.Ctx)error{
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		client, _ := mongo.Connect(ctx, options.Client().ApplyURI(url))
		id,_ := primitive.ObjectIDFromHex(c.Params("id"))
		filter :=bson.M{"_id":	id}
		collection:= client.Database("test").Collection("todos")
		 _,err := collection.DeleteOne(ctx, filter)
		if err!=nil {
			log.Fatal(err)
		}
		defer client.Disconnect(ctx)
		return c.Status(201).SendString("Todo Deleted")

	})
	app.Get("/about", func(c *fiber.Ctx) error {
		return c.SendString("Hello From Golang Fiber: INSPIRED BY EXPRESS.JS")
	})



	app.Static("_next/static","./frontend/.next/static")
	app.Static("/vercel.svg","./frontend/public/vercel.svg")
	app.Static("/favicon.ico","./frontend/public/favicon.ico")
	app.Static("/react-js", "./frontend/.next/server/pages/index.html")
	PORT :=":8080"
	app.Listen(PORT)
}
