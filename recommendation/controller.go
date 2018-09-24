package recommendation

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Controller struct{}

func (c *Controller) getPillsForDeficit(w http.ResponseWriter, r *http.Request) {
	var defictits Deficits
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576)) // read the body of the request
	if err != nil {
		log.Fatalln("Error getPillsForDeficit", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := r.Body.Close(); err != nil {
		log.Fatalln("Error getPillsForDeficit", err)
	}

	if err := json.Unmarshal(body, &defictits); err != nil { // unmarshall body contents as a type Candidate
		w.WriteHeader(422) // unprocessable entity
		log.Println(err)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			log.Fatalln("Error getPillsForDeficit unmarshalling data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	w.WriteHeader(http.StatusOK)
	pillInventoriesChan := make(chan *PillInventories)
	go getPillInventory(pillInventoriesChan, &defictits)
	inventory := <-pillInventoriesChan
	optimizeChannel := make(chan *Pills)
	go optimize(optimizeChannel, inventory, &defictits)
	pills := <-optimizeChannel
	if len(*pills) > 0 {
		data, err := json.Marshal(pills)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	} else {
		fmt.Println("No data")
		w.WriteHeader(http.StatusNoContent)
	}
	return
}
