package main

/*
  shipping page html

  This Lambda function serves the shipping.html file to the end-user during the checkoout process.
  The shipping.html file is stored locally in the Lambda function instance in a Zip folder
  next to the shipping.go binary file (Serverless Function Code Uri).

  Shipping rates for the customer's order are calculated in a PUT call to getShippingMethods
  and returned as options to the user. The order is updated with the selected rate once
  the user proceeds to the payment page.

*/

import (
	"log"
	"net/http"

	"github.com/apex/gateway"
	"github.com/ggarcia209/acamoprjct/service/util/httpops"
)

const route = "/store/checkout/shipping" // GET
const path = "./shipping.html"

// RootHandler handles HTTP request to the root '/'
func RootHandler(w http.ResponseWriter, r *http.Request) {
	httpops.HtmlHandler(w, r, path)
}

func main() {
	httpops.RegisterRoutesHtml(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
