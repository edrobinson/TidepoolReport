/*
   This package implements the Tidepool APIs to get authorization
   then user data for blood glucose values using Golang.

   There are only a few functions of consequence:

   1. home - renders the "home page" that takes request parameters from the user as:
       1. User email
       2. Password
       3. Optional Starting date
       4. Optional Ending date
       5. Optional data type. Defaults to "smbg" or regular glucose readings taken by the user.
          NOTE: Only smbg supported now... 5/13/21

       This only supports the smbg (S_elf M_anaged B_lood G_lucoses - finger sticks) type initially

   2. send - receives the options data from the browser and runs the api calls to Tidepool to retrieve
             authorization followed by a data request using HTTP requests. The data returned
             is a JSON string which is saved to a .json file for further processing.

   3. CreatePDF - called from the API processor, this function uses the gofpdf package to create and
      store a PDF of the Glucose values.

   4. ShowPDF - displays the pdf in the browser.
*/

package tidepoolreport

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"
    "errors"
) 

//Tidepool error response message.
//For things like 403 errors when user enters invalid credentials
type tpError struct {
    Status      int
    Id          string  
    Code        string
    Message     string  
}


//Tidepool structures generated by github JSONtoGo
type tpMeasurement []struct {
	Conversionoffset    int           `json:"conversionOffset"`
	Deviceid            string        `json:"deviceId"`
	Devicetime          string        `json:"deviceTime"`
	GUID                string        `json:"guid"`
	ID                  string        `json:"id"`
	Payload             Payload       `json:"payload,omitempty"`
	Time                time.Time     `json:"time"`
	Timezoneoffset      int           `json:"timezoneOffset"`
	Type                string        `json:"type"`
	Units               string        `json:"units,omitempty"`
	Uploadid            string        `json:"uploadId"`
	Value               float64       `json:"value,omitempty"`
	Annotations         []Annotations `json:"annotations,omitempty"`
	Byuser              string        `json:"byUser,omitempty"`
	Client              Client        `json:"client,omitempty"`
	Computertime        string        `json:"computerTime,omitempty"`
	Devicemanufacturers []string      `json:"deviceManufacturers,omitempty"`
	Devicemodel         string        `json:"deviceModel,omitempty"`
	Deviceserialnumber  string        `json:"deviceSerialNumber,omitempty"`
	Devicetags          []string      `json:"deviceTags,omitempty"`
	Timeprocessing      string        `json:"timeProcessing,omitempty"`
	Timezone            string        `json:"timezone,omitempty"`
	Version             string        `json:"version,omitempty"`
}

//Additional structures passed by Tidepool
//inside the measurement structure.
//This code does not use them

//Payload - not used
type Payload struct {
	Logindices []int `json:"logIndices"`
}

//Annotations - not used
type Annotations struct {
	Code string `json:"code"`
}

//Private - not used
type Private struct {
	Os string `json:"os"`
}

//Client - not used
type Client struct {
	Name    string  `json:"name"`
	Private Private `json:"private"`
	Version string  `json:"version"`
}

//This is the structure passed to the PDF generator
//Date, time and value
type Smbg struct {
	smbgDate  string
	smbgTime  string
	smbgValue string
}




// Simple error checking - not too friendly
func check(e error, msg string) {
	if e != nil {
		log.Fatal(msg, e)
	}
}


//Set up routing and start the web server
func main() {

    http.Handle("/", http.HandlerFunc(home))     //Serve the home page
	http.Handle("/opts", http.HandlerFunc(send)) //Run the Tidepool api and gen the pdf of the results

	//Serve statics like css and js - see the static folder.
    //Took me a lot of time to get this straight...
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Listening... Go to localhost:3000")
	
    err := http.ListenAndServe(":3000", nil) //Start a server instance and Listen on port 3000
	check(err, "Error on server start")      //Oops...
}

//Render the home screen with options form
func home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/TidepoolMain.html")
	check(err, "Can't parse main template.")
	tmpl.Execute(w, nil)
}

/*
   1. Receive the request from the browser with the form data.
   2. Parse the form
   3. Access Tidepool for authorization sending the users id (Email)
      and password with a POST request
   4. Retrieve the auth token from the response header.
   5. Retrieve the userid from the response body.
   6. Make a GET call to retrieve the user data.
   7. Call for the PDF generator.
   8. Show the PDF in the browser.
*/
func send(w http.ResponseWriter, r *http.Request) {
	//Get the form values from the response
	r.ParseForm()

	/*
	   The first step is to get authorization from Tidepool
	   using our Tidepool user id (Email) and password
	*/
	//Create a POST request to the Tidepool authorization api
	req, err := http.NewRequest("POST", "https://int-api.tidepool.org/auth/login", nil)
	check(err, "Error creating the auth request")

	//Use basic uid/pwd authentication
	req.SetBasicAuth(r.PostFormValue("useremail"), r.PostFormValue("password"))

	//Send the request
	resp, err := http.DefaultClient.Do(req)
	check(err, "Error sending the auth request")
	defer resp.Body.Close()

	//Not OK response?
	if resp.StatusCode != 200 {
		check(nil, "Authorization Call: Bad request response"+resp.Status)
	}

	//Get the Tidepool token header from the response headers
	var token = resp.Header.Get("x-tidepool-session-token")

	//Get the Tidepool user account id from the json response body
	//1. Read the response body
	bytes, err := ioutil.ReadAll(resp.Body)
	check(err, "Error reading the auth response body")

	//2. Decode the json string into a map
	var result map[string]interface{}
	json.Unmarshal([]byte(bytes), &result) //Unmarshal into the result map

	//3. Get the user id from the body map
	var userid = fmt.Sprintf("%v", result["userid"])

	/*
	   At this point we have the credentials we need to request the users data
	   We'll setup and make a GET request to the data api.
	*/
	

	//The url contains the Tidepool internal userid for the login.
    //The url is asking for finger stick measurements - ?type=smbg.
	var url string = "https://int-api.tidepool.org/data/" + userid + "?type=" + r.PostFormValue("datatype")

	//Add the start and/or end dates to the query string.
	var queryString string = checkDateRanges(r.PostFormValue("startdate"), r.PostFormValue("enddate"))
	if queryString != "" {
		url = url + queryString
	}

	//Instance a GET request
	req, err = http.NewRequest("GET", url, nil)
	check(err, "Error creating data request")

	//Set the headers - token and content type
	req.Header.Set("x-tidepool-session-token", token)
	req.Header.Set("content-type", "application/json")

	//Execute the request
	resp, err = http.DefaultClient.Do(req)
	check(err, "Error executing data request")

	defer resp.Body.Close()

	//Check the http respose code - want 200 OK
	if resp.StatusCode != 200 {
		check(nil, "Data API call: Unexpected response status =  "+resp.Status)
	}

	//Get the body of the response - contains the requested test results
	//data is read as byte type
	data, err := ioutil.ReadAll(resp.Body)
	check(err, "Error getting the data file from response body.")

	//Write it to a file
	err = ioutil.WriteFile("tidepool.json", data, 0775)
	check(err, "Error saving the result data file")

    
    //Extract the result data
    err, s := decodeTidepoolData("tidepool.json")
    if err != nil{
        _ = CheckTidepoolErrorResponse(w,"tidepool.json") //Handle tidepool things like 403 error
        return
    }
    
    //Empty result set?
    if len(s) == 0 {
        log.Println("No results were returned from Tidepool.")
    }

    CreatePDF(w, s)

	//Display the pdf in the browser
	ShowPDF(w, r, "tidepool.pdf")
}

/*
   The user optionally enters a start date and/or end date of results to be returned.
   This function evaluates these form inputs and returns
   a query string or an empty string.

   The inputs are of form yyyy-mm-dd
   Tidepool wants them in this format 2015-10-10T15:00:00.000Z
   This works out well as we do not have to mess with any of the time functions.
*/
func checkDateRanges(sdate string, edate string) string {
	var qs string = "" //Initial an empty query string

	if sdate == "" && edate == "" {
		return qs
	} //No dates entred

	var datetail string = "T01:00:00.000Z" //The time portion of the dt string

	if sdate != "" {
		qs = qs + "&startDate=" + sdate + datetail
	}
	if edate != "" {
		qs = qs + "&endDate=" + edate + datetail
	}
	return qs
}

//Extract the result fields into s slice of smbg structs
func decodeTidepoolData(filename string) (error, []Smbg){
	var smbgs []Smbg //Slice of smbg structures
	var psmbg Smbg //An smbg struct object

	//Load the result set
    file, err := ioutil.ReadFile(filename)
	check(err, "Error loading result json file")

	//Tidepool smbg struct 
    result := tpMeasurement{}
    
	//Extract the measurement records
    err = json.Unmarshal([]byte(file), &result)
    if err != nil{
        return errors.New("Tidepool appears to have returned an error response"), nil
    }
    
	//Scan the json and construct the smbg array to pass to the pdf writer.
	//All we pass is date, time and value in a structure of strings
	for i := range result {
		//The smbg type is the measurement we want. A few others show up...
        if result[i].Type != "smbg" {
			continue
		} 

		//Break out the measurement date & time
		var measdt string = result[i].Devicetime //Example: 2021-03-17T08:33:00
		var measDate string = measdt[:10]        //Date string
		var measTime string = measdt[11:19]      //Time string

		//The test result arrives as a float representing Mmols/L. We want mg/dl
		//Conversion is Mmol/L * 18 = mg/dl. We want an integer string
		var measvals string = strconv.Itoa(int(result[i].Value * 18)) //To mg/dl -> integer -> string

		//Fill out the smbg structure
		psmbg.smbgDate = measDate
		psmbg.smbgTime = measTime
		psmbg.smbgValue = measvals

		//Append it to the smbg slice
		smbgs = append(smbgs, psmbg)
        
	}
    return nil, smbgs
    
}


//Load and Render the HTML to the browser.
//Called by the router events in main().
func render(w http.ResponseWriter, filename string, data interface{}) {
	//Load and parse the html file
	tmpl, err := template.ParseFiles(filename)
	if err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
	}

	//Output the html to the browser
	if err := tmpl.Execute(w, data); err != nil {
		log.Println(err)
		http.Error(w, "Sorry, something went wrong", http.StatusInternalServerError)
	}
}

