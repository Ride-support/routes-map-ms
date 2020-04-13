package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/umahmood/haversine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

// Cardinal is the database structure
type Cardinal struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Latitude  float64            `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude float64            `json:"longitude,omitempty" bson:"longitude,omitempty"`
}

// Distance is the heversine distance betwhen 2 coordinates
type Distance struct {
	ID primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Km float64            `json:"km,omitempty" bson:"km,omitempty"`
}

// CreateCoordinateEndpoint is the method that creates new points in map
func CreateCoordinateEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var coordinate Cardinal
	_ = json.NewDecoder(request.Body).Decode(&coordinate)
	collection := client.Database("db").Collection("coordinates")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()
	result, _ := collection.InsertOne(ctx, coordinate)
	json.NewEncoder(response).Encode(result)
}

// DeleteCoordinateEndpoint find coordinate by id and delete it
func DeleteCoordinateEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var coordinate Cardinal
	collection := client.Database("db").Collection("coordinates")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := collection.FindOneAndDelete(ctx, Cardinal{ID: id}).Decode(&coordinate)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(coordinate)
}

// GetCoordinatesEndpoint retuns all the points in existence
func GetCoordinatesEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	var coordinates []Cardinal
	collection := client.Database("db").Collection("coordinates")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var person Cardinal
		cursor.Decode(&person)
		coordinates = append(coordinates, person)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(coordinates)
}

// GetCoordinateEndpoint recieve the id of coordinate and return the structure
func GetCoordinateEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var coordinate Cardinal
	collection := client.Database("db").Collection("coordinates")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	err := collection.FindOne(ctx, Cardinal{ID: id}).Decode(&coordinate)
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(coordinate)
}

// GetDistancesEndpoint calculate the distance betwhen user corrdinate vs all of the other coordinates
func GetDistancesEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("content-type", "application/json")
	params := mux.Vars(request)
	llat, _ := strconv.ParseFloat(params["lat"], 64)
	llon, _ := strconv.ParseFloat(params["lon"], 64)
	local := haversine.Coord{Lat: llat, Lon: llon}
	var distances []Distance
	collection := client.Database("db").Collection("coordinates")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var coordinate Cardinal
		var distance Distance
		cursor.Decode(&coordinate)
		parking := haversine.Coord{Lat: coordinate.Latitude, Lon: coordinate.Longitude}
		_, km := haversine.Distance(local, parking)
		distance.ID = coordinate.ID
		distance.Km = km
		distances = append(distances, distance)
	}
	if err := cursor.Err(); err != nil {
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{ "message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(distances)
}

func home(response http.ResponseWriter, request *http.Request) {

	fuck := "works fine"
	json.NewEncoder(response).Encode(fuck)

}

func main() {
	fmt.Println("Starting the application...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	clientOptions := options.Client().ApplyURI("mongodb://routes-map-ms_routes-map-db_1:27017")
	client, _ = mongo.Connect(ctx, clientOptions)
	router := mux.NewRouter()
	router.HandleFunc("/coordinate", CreateCoordinateEndpoint).Methods("POST")
	router.HandleFunc("/coordinate/{id}", DeleteCoordinateEndpoint).Methods("DELETE")
	router.HandleFunc("/coordinates", GetCoordinatesEndpoint).Methods("GET")
	router.HandleFunc("/coordinate/{id}", GetCoordinateEndpoint).Methods("GET")
	router.HandleFunc("/", home).Methods("GET")
	router.HandleFunc("/distances/{lat}/{lon}", GetDistancesEndpoint).Methods("GET")
	http.ListenAndServe(":9090", router)
}
