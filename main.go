package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	// define handlers
	http.Handle("/", http.FileServer(http.Dir("./static")))
	http.HandleFunc("/analysis", func(w http.ResponseWriter, r *http.Request) {
		// check http method
		if r.Method != http.MethodPost {
			WriteAPIResp(w, NewErrorResp(NewErrMethodNotAllowed()))
			return
		}
		// parse request body
		var reqBody analyzeReqBody
		err := json.NewDecoder(r.Body).Decode(&reqBody)
		if err != nil {
			WriteAPIResp(w, NewErrorResp(NewErrBadRequest(err.Error())))
			return
		}
		// validate request body
		err = reqBody.Validate()
		if err != nil {
			WriteAPIResp(w, NewErrorResp(err))
			return
		}
		// do analysis
		matches := doAnalysis(reqBody.InputText, reqBody.RefText)
		// output success response
		WriteAPIResp(w, NewSuccessResp(map[string]interface{}{
			"matches": matches,
		}))
	})
	// define port, we need to set it as env for Heroku deployment
	port := os.Getenv("PORT")
	if port == "" {
		port = "9056"
	}
	// run server
	log.Printf("server is listening on :%v", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		log.Fatalf("unable to run server due: %v", err)
	}
}

func doAnalysis(input, ref string) []match {
	splitInput := strings.SplitAfter(input, " ")
	splitRef := strings.SplitAfter(ref, " ")
	splitDiff := compareDiff(input, splitRef)
	splitEqual := compareEqual(input, splitRef)
	check := false
	
	jsonInput, errInput := json.Marshal(splitInput)
	jsonRef, errRef := json.Marshal(splitRef)
	jsonComp, errComp := json.Marshal(splitEqual)
	
	i := 0
	inputFirst := 0
	for _, a := range splitInput {
		if inputFirst <= 0 {
			inputFirst = strings.Index(ref, a)
		}
		i++
	}

	i = len(splitInput) - 1
	inputLast := 0
	for _, b := range splitInput {
		if inputLast <= 0 {
			inputLast = strings.Index(ref, b)
		}
		i--
	}
	
	anyFirstWords := "#"
	if (inputFirst > 0) {
		anyFirstWords = ref[:inputFirst]
	}

	anyLastWords := "#"
	if (inputLast > 0) {
		anyLastWords = strings.Trim(ref, ref[:inputLast])
	}

	// check trimmed pharse structure
	if len(strings.SplitAfter(anyFirstWords, " ")) > 4 || len(strings.SplitAfter(anyLastWords, " ")) > 4 {
		if strings.Index(strings.Join(splitDiff, " "), anyFirstWords) == -1 || strings.Index(strings.Join(splitDiff, " "), anyLastWords) == -1  {
			check = true
		}
	}

	// limit same pharse
	if (len(splitEqual) <= 4 ) {
		check = false
	}
	
	return []match{
		{
			Input: matchDetails{
				Text:     input,
				StartIdx: 0,
				EndIdx:   len(input) - 1,
			},
			Reference: matchDetails{
				Text:     ref,
				StartIdx: 0,
				EndIdx:   len(ref) - 1,
			},
			Json: jsonRes{
				TextRef:   string(jsonRef),
				TextInput: string(jsonInput),
				TextComp:  string(jsonComp),
				ErrRef:    errRef,
				ErrInput:  errInput,
				ErrComp:   errComp,
			},
			Result: check,
		},
	}
}

func compareDiff(s string, t []string) []string {
	doCompare := func(x string, y string) int { 
		return strings.Index(strings.ToLower(strings.Trim(x, " ")), strings.ToLower(strings.Trim(y, " "))) }
	var r []string
	i := 0
	isEqual := true
	for _, u := range t {
		isEqual = false
		if doCompare(s, u) == -1 {
			isEqual = true
		}
		if (isEqual) {
			r = appendDiff(r, strings.ToLower(strings.Trim(t[i], " ")))
		}
		i++
	}
	return r
}

func compareEqual(s string, t []string) []string {
	doCompare := func(x string, y string) int { 
		return strings.Index(strings.ToLower(strings.Trim(x, " ")), strings.ToLower(strings.Trim(y, " "))) }
	var r []string
	i := 0
	isEqual := true
	for _, u := range t {
		isEqual = false
		if doCompare(s, u) == -1 {
			isEqual = true
		}
		if (!isEqual) {
			r = appendDiff(r, strings.ToLower(strings.Trim(t[i], " ")))
		}
		i++
	}
	return r
}

func appendDiff(slice []string, i string) []string {
    for _, ele := range slice {
        if ele == i {
            return slice
        }
    }
    return append(slice, i)
}
