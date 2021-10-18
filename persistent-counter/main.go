package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/vorteil/direktiv-apps/pkg/direktivapps"
)

const code = "com.%s.error"

type counterAPIResp struct {
	Value    int `json:"value"`
	OldValue int `json:"old_value"`
}

// request the input object for the requester container
type request struct {
	UUID string `json:"uuid"`
	Min  int    `json:"min"`
}

// output for the requester container
type output struct {
	Value int `json:"value"`
}

func Request(w http.ResponseWriter, r *http.Request) {
	obj := new(request)
	aid, err := direktivapps.Unmarshal(obj, r)
	if err != nil {
		direktivapps.RespondWithError(w, fmt.Sprintf(code, "unmarshal-input"), err.Error())
		return
	}

	direktivapps.Log(aid, "Creating new counter request")
	resp, err := http.Get(fmt.Sprintf("https://api.countapi.xyz/hit/%s/key", obj.UUID))
	if err != nil {
		direktivapps.RespondWithError(w, fmt.Sprintf(code, "request"), err.Error())
		return
	}

	direktivapps.Log(aid, "Processing counter request")
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		direktivapps.RespondWithError(w, fmt.Sprintf(code, "read-body"), err.Error())
		return
	}

	apiResp := new(counterAPIResp)
	err = json.Unmarshal(body, apiResp)
	if err != nil {
		direktivapps.RespondWithError(w, fmt.Sprintf(code, "marshal-body"), err.Error())
		return
	}

	direktivapps.Log(aid, "Preparing response")
	if obj.Min != 0 && apiResp.Value < obj.Min {
		direktivapps.RespondWithError(w, fmt.Sprintf(code, "invalid-value"), fmt.Sprintf("counter value=%v is less than min=%v", apiResp.Value, obj.Min))
		return
	}

	var responding output
	responding.Value = apiResp.Value

	data, err := json.Marshal(responding)
	if err != nil {
		direktivapps.RespondWithError(w, fmt.Sprintf(code, "marshal-output"), err.Error())
		return
	}

	direktivapps.Respond(w, data)
}

func main() {
	direktivapps.StartServer(Request)
}
