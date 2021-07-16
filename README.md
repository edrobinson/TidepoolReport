# TidepoolReport
 Tidepool.org dibetic data report generator
 
Tidepool.org is a non-profit thatallows diabetics to upload their diabetes information to be stored on the Tidepool system. Once stored the users can view several reports and they can download them and share them with their healthcare providers. Tidepool does not, however, provide a simple list of date, time and value for blood glucose values.
 
This go project queries the Tidepool data apis to retrieve credentials and test results. The results are output to a PDF as a simple 3 column report.

Usage:
1. Download or Clone the project.
2. In your command line tool issue go build.
3. Enter tidepoolreport or ./tidepoolreport if not using windoze.
4. Go to localhost:3000, fill in the options - you must have a regular tidepool account-, optionally select a date range and only select the SMBG type.
5. Submit the form and the pdf should appear with blinding speed. :)

As presented, this project queries the Tidepool development servers. 

 
