package tidepoolreport

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"net/http"
    "errors"
)


//CheckTidepoolErrorResponse attempte to decode the Tidepool response body.
//Assuming it is an error response because i could not be decoded as a  result set
func CheckTidepoolErrorResponse(w http.ResponseWriter, filename string) (err error){
    var tpe  tpError

    //Load the result set
    file, err := ioutil.ReadFile(filename)
	check(err, "Error loading result json file")
    err = json.Unmarshal([]byte(file), &tpe)
    if err != nil{
        return errors.New("Unable to decode assumed Tidepool error response.")
    }
    
    tmpl, err := template.ParseFiles("templates/ErrorMessageScreen.html")
    check(err, "Failed to parse the error message template.")
    
    err =  tmpl.Execute(w, tpe)
    check(err, "Failed to execute the error response template")
    
    return errors.New("Displayed the Tidepool Error Page.")
}

//DisplayMessageScreen - general purpose messager.
//Param is a single string. 
func DisplayMessageScreen(w http.ResponseWriter, msg string){
            
        tmpl, err := template.ParseFiles("templates/ErrorMessageScreen.html")
        check(err, "Failed to parse the error message template.")
        
        err =  tmpl.Execute(w, msg)
        check(err, "Failed to execute the error  template")
}
