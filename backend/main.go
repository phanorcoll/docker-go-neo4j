package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

//Neo4jConfiguration holds the configuration for connecting to the DB
type Neo4jConfiguration struct {
	URL      string
	Username string
	Password string
	Database string
}

//MovieResult ...
type MovieResult struct {
	Movie `json:"movie"`
}

//Movie ...
type Movie struct {
	Released int64  `json:"released"`
	Title    string `json:"title,omitempty"`
}

//Person will structure the data for the Person records
type Person struct {
	Job  string   `json:"job"`
	Role []string `json:"role"`
	Name string   `json:"name"`
}

//newDrive is a method for Neo4jConfiguration to return a connection to the DB
func (nc *Neo4jConfiguration) newDriver() (neo4j.Driver, error) {
	return neo4j.NewDriver(nc.URL, neo4j.BasicAuth(nc.Username, nc.Password, ""))
}

//getDataFunc will query the database and return the result as JSON
func getDataFunc(driver neo4j.Driver) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		session := driver.NewSession(neo4j.SessionConfig{
			AccessMode:   neo4j.AccessModeRead,
			DatabaseName: "neo4j",
		})
		defer unsafeClose(session)

		query := `MATCH (movie:Movie) RETURN movie.title as title, movie.released as released LIMIT $limit`
		result, err := session.Run(query, map[string]interface{}{"limit": 10})
		if err != nil {
			log.Println("Error querying Neo4j", err)
			return
		}
		var movieResults []MovieResult
		for result.Next() {
			record := result.Record()
			released, _ := record.Get("released")
			title, _ := record.Get("title")
			movieResults = append(movieResults, MovieResult{Movie{
				Released: released.(int64),
				Title:    title.(string),
			}})
		}

		moviesJSON, err := json.Marshal(movieResults)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Write(moviesJSON)

	}
}

func main() {
	configuration := parseConfiguration()
	driver, err := configuration.newDriver()
	if err != nil {
		log.Fatal(err)
	}
	defer unsafeClose(driver)
	http.HandleFunc("/", getDataFunc(driver))
	http.ListenAndServe(":8080", nil)
}

func parseConfiguration() *Neo4jConfiguration {
	return &Neo4jConfiguration{
		URL:      "neo4j://neo4j:7687",
		Username: "neo4j",
		Password: "testing",
	}
}

func unsafeClose(closeable io.Closer) {
	if err := closeable.Close(); err != nil {
		log.Fatal(fmt.Errorf("could not close resource: %w", err))
	}
}
