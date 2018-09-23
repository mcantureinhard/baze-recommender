package recommendation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func getPillInventory(respond chan<- *PillInventories, deficits *Deficits) {
	url := "http://localhost:8081/getPillsWithMicroNutrients"
	jsonData, err := json.Marshal(deficits)
	if err != nil {
		panic(err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("X-Custom-Header", "myvalue")
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	var pillInventories PillInventories
	if err := json.Unmarshal(body, &pillInventories); err != nil { // unmarshall body contents as a type Candidate
		respond <- &pillInventories
		fmt.Println(err)
	}
	respond <- &pillInventories
}
