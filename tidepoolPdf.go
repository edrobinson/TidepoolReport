package tidepoolreport

import (
	"bytes"
	//"encoding/json"
	"fmt"
	"github.com/jung-kurt/gofpdf"
	//"html/template"
	"io/ioutil"
	//"log"
	"net/http"
	"os"
	//"strconv"
	//"time"
    //"errors"
)


//Setup the pdf generator
var pdf = gofpdf.New("P", "in", "letter", "") //portrait, inches, letter size

/*
   Using the gofpdf package, create a pdf file from the
   users measurments data
   The filename param is the file that contains the downloaded json.
   The pdf ge. object is instanced up top for global access
*/
func CreatePDF(w http.ResponseWriter, smbgs []Smbg) error{

	/*
	   Now we are ready to produce the PDF.
	   Initially I am creating a pretty basic PDF
	   with no fancy page headings, etc.
	   Stay tuned...
	*/

	//Set up the page header function - kind of an override...
	pdf.SetHeaderFunc(func() {
		pdf.SetY(.2)
		pdf.SetFont("Arial", "B", 15)
		//pdf.Cell(2.2, 0, "")
		pdf.CellFormat(0, .4, "Glucose Values", "", 0, "C", false, 0, "")
		pdf.Ln(.5)
		//Add the column headers
		lineOut("Date", "Time", "Glucose mg/dl")

	})

	//Set the page footer function.
	pdf.SetFooterFunc(func() {
		pdf.SetY(-.5)
		pdf.SetFont("Arial", "I", 8)
		pdf.CellFormat(0, .4, fmt.Sprintf("Page %d /{nb}", pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})

	pdf.AliasNbPages("")         //Gets us page/pages in the footer
	pdf.AddPage()                //Put in the first page
	pdf.SetFont("Arial", "", 12) //Set the document font

	//Add all of the measurements.
	for i := range smbgs {
		lineOut(smbgs[i].smbgDate, smbgs[i].smbgTime, smbgs[i].smbgValue)
	}

	//Store the pdf file and cleanup.
	pdf.OutputFileAndClose("tidepool.pdf")
    return nil
}

//Output a result line of cells to the pdf.
func lineOut(s1, s2, s3 string) {
	pdf.Cell(1.35, 0, "") //1" indent
	cellOut(s1)
	cellOut(s2)
	cellOut(s3)
	pdf.Ln(0.3) //End of line
}

//Standardize the cell format.
func cellOut(s string) {
	pdf.CellFormat(1.7, 0.3, s, "1", 0, "C", false, 0, "")
}

//Render the pdf to the browser.
func ShowPDF(w http.ResponseWriter, r *http.Request, filename string) {
	//Load the PDF file
	streamPDFbytes, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	//To buffer
	b := bytes.NewBuffer(streamPDFbytes)

	//Let 'em know what's coming
	w.Header().Set("Content-type", "application/pdf")

	//Write the file bytes to the brower
	if _, err := b.WriteTo(w); err != nil {
		fmt.Fprintf(w, "%s", err)
	}
}
