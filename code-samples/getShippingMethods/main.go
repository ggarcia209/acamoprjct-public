package main

/* getShippingMethods retrieves the available shipping methods for the current active order.
   Shipping rates are caculated by determining the total weight of the roder, and how many packages
   are required to ship the order, using a greedy algorithm that optimizes for price. The price per
   cubic inch decreases as the total volume of the shipping parcel increases. Given this, price is
   optimized by using the smallest available package that will fit the whole order. If multiple
   packages are required, the largest available parcel will be used to package the order, until
   there is a smaller parcel that can fit the remaining order volume.
*/

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/apex/gateway"
	"github.com/coldbrewcloud/go-shippo"
	"github.com/coldbrewcloud/go-shippo/client"
	"github.com/coldbrewcloud/go-shippo/models"
	"github.com/ggarcia209/acamoprjct/service/store-api/store"
	"github.com/ggarcia209/acamoprjct/service/util/dbops"
	"github.com/ggarcia209/acamoprjct/service/util/httpops"
	"github.com/ggarcia209/acamoprjct/service/util/shipops"
	"github.com/ggarcia209/acamoprjct/service/util/sortops"
	"github.com/ggarcia209/go-aws/go-dynamo/dynamo"
)

const route = "/store/checkout/get_rates" // PUT

const failMsg = "Request failed!"
const successMsg = "Request succeeded!"

// list of tables function makes r/w calls to
var tables = []dbops.Table{
	dbops.Table{
		Name:       dbops.CustomersTable(),
		PrimaryKey: dbops.CustomersPK,
	},
	dbops.Table{
		Name:       dbops.OrdersTable(),
		PrimaryKey: dbops.OrdersPK,
		SortKey:    dbops.OrdersSK,
	},
	dbops.Table{
		Name:       dbops.StoreItemsIndexTable(),
		PrimaryKey: dbops.StoreItemsIndexPK,
	},
	dbops.Table{
		Name:       dbops.ParcelsTable(),
		PrimaryKey: dbops.ParcelsPK,
	},
	dbops.Table{
		Name:       dbops.ShipmentsTable(),
		PrimaryKey: dbops.ShipmentsPK,
	},
}

// customerInfo represents the request info submitted from the /store/checkout/shipping page
type customerInfo struct {
	UserID  string `json:"user_id"`
	OrderID string `json:"order_id"`
}

// getDimensions() return type
type dimensions struct {
	Weight    float32
	Volume    float32
	MaxLength float32
	MaxWidth  float32
	MaxHeight float32
}

// Box represents a 3D cubic space and is a subcomponent of the dynamic programming model
// for filling parcels with each item in an Order. Each filled Box contains 1 PkgItem
// bordered by 3 child Box nodes along the PkgItem's X, Y, & Z axes, derived from its remaining 3D space.
// The Root Box Node represents the selected shipping Parcel to fill with CartItem units
// and the whole of it's volume.
type Box struct {
	Volume  float32
	ResvPct float32 // percentage of parcel volume reserved for packing materials
	Length  float32
	Width   float32
	Height  float32
	Item    string
	NodeL   *Box
	NodeW   *Box
	NodeH   *Box
}

// PkgItem represents an individual unit of a CartItem in an Order and the 3D space it occupies within a Box / Parcel (Box Root Node).
// PkgItems are sorted by volume greatest to least, and are successively added to the child Nodes of a Box and their descendants
// as remaining space permits.
type PkgItem struct {
	ItemID string
	Name   string
	Volume float32
	Length float32
	Width  float32
	Height float32
}

// Add adds an item to the current Box, and creates 3 child Box nodes.
// Each node represents the remaining space derived from the current Box
// in the form of a smaller, empty Box formed along the X, Y, and Z axes of the item,
// bounded by the dimensions of the current, occupied Box.
// Each smaller box is recursively filled with the next largest item until there are no remaining items,
// or there is an insufficient amount of space in the Root Box (node) for the remaining items.
func (b *Box) Add(item PkgItem) error {
	ok := true
	resv := b.ResvPct + 1.0
	// compare along y axis orientation (default)
	if item.Length > (b.Length/resv) || item.Width > (b.Width/resv) || item.Height > (b.Height/resv) {
		ok = false
	}
	if !ok {
		// compare along x axis orientation
		if item.Width > (b.Length/resv) || item.Length > (b.Width/resv) || item.Height > (b.Height/resv) {
			ok = false
		} else {
			ok = true
		}
	}
	if !ok {
		// compare along z axis orientation
		if item.Height > (b.Length/resv) || item.Length > (b.Width/resv) || item.Length > (b.Height/resv) {
			ok = false
		} else {
			ok = true
		}
	}
	if !ok {
		// insufficient space
		return fmt.Errorf("DIMENSIONS_EXCEEDED")
	}

	// add item to current box; create child nodes
	b.Item = item.ItemID

	// l = (Lx - Ly, Wy, Hx)
	lBox := &Box{
		Length: b.Length - item.Length,
		Width:  item.Width,
		Height: b.Height,
	}
	lBox.Volume = lBox.Length * lBox.Width * lBox.Height

	// w = (Lx, Wx - Wy, Hx)
	wBox := &Box{
		Length: b.Length,
		Width:  b.Width - item.Width,
		Height: b.Height,
	}
	wBox.Volume = wBox.Length * wBox.Width * wBox.Height

	// h = (Ly, Wy, Hx - Hy)
	hBox := &Box{
		Length: item.Length,
		Width:  item.Width,
		Height: b.Height - item.Height,
	}
	hBox.Volume = hBox.Length * hBox.Width * hBox.Height

	b.NodeL, b.NodeW, b.NodeH = lBox, wBox, hBox

	return nil
}

// RootHandler handles HTTP request
func RootHandler(w http.ResponseWriter, r *http.Request) {
	// DB is used to make DynamoDB API calls
	DB := dbops.InitDB(tables)

	// verify content-type
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		httpops.ErrResponse(w, "Content-Type is not application/json", failMsg, http.StatusUnsupportedMediaType)
		return
	}

	// decode JSON object from http request
	data := customerInfo{}
	var unmarshalErr *json.UnmarshalTypeError

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	err := decoder.Decode(&data)
	if err != nil {
		if errors.As(err, &unmarshalErr) {
			httpops.ErrResponse(w, "Bad Request: Wrong type provided for field "+unmarshalErr.Field, failMsg, http.StatusBadRequest)
		} else {
			httpops.ErrResponse(w, "Bad Request: "+err.Error(), failMsg, http.StatusBadRequest)
		}
		return
	}

	if data.UserID == "" || data.OrderID == "" {
		log.Printf("bad request - empty keys")
		httpops.ErrResponse(w, "Bad Request: empty order keys", failMsg, http.StatusBadRequest)
		return
	}

	// initialize shippo client
	token, err := getToken()
	if err != nil {
		log.Printf("RootHandler failed - getToken: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}
	c := shippo.NewClient(token)

	// get order items
	order, err := dbops.GetOrder(DB, data.UserID, data.OrderID)
	if err != nil {
		log.Printf("RootHandler failed - getOrder: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	// get shipping rates
	rates, shipment, err := getShippingRates(DB, c, data, order)
	if err != nil {
		log.Printf("RootHandler failed - getShippingRates: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), failMsg, http.StatusInternalServerError)
		return
	}

	sorted := sortops.SortRatesByPrice(rates)
	shipment.Rates = []store.RateSummary{}
	for _, rate := range sorted {
		shipment.Rates = append(shipment.Rates, *rate)
	}

	// create shipment in DB
	err = dbops.PutShipment(DB, &shipment)
	if err != nil {
		log.Printf("RootHandler failed - putShipment: %v", err)
		httpops.ErrResponse(w, "Internal Server Error: "+err.Error(), "SAVE_SHIPPING_ADDRESS_FAIL", http.StatusInternalServerError)
		return
	}

	// return shipping rates
	httpops.ErrResponse(w, "Shipping rates: ", sorted, http.StatusOK)
	return
}

// get shippo API token from disk
func getToken() (string, error) {
	token, err := shipops.GetToken("./stk.txt")
	if err != nil {
		log.Printf("getToken failed: %v", err)
		return "", err
	}
	return token, nil
}

// get shipping rates for order
func getShippingRates(DB *dynamo.DbInfo, c *client.Client, data customerInfo, order *store.Order) ([]store.RateSummary, store.Shipment, error) {
	// create to/from addresses
	to, err := createShipmentAddress(c, order.ShippingAddress)
	if err != nil {
		log.Printf("getShippingRates failed: %v", err)
		return nil, store.Shipment{}, err
	}

	from, err := createReturnAddress(c)
	if err != nil {
		log.Printf("getShippingRates failed: %v", err)
		return nil, store.Shipment{}, err
	}

	// create parcels
	parcelIDs, err := dbops.GetStoreItemIndex(DB, "parcels-"+store.CarriersUsps) // fixed to USPS
	if err != nil {
		log.Printf("getShippingRates failed: %v", err)
		return nil, store.Shipment{}, err
	}

	parcelObjs, err := dbops.BatchGetParcels(DB, parcelIDs.ItemIDs)
	if err != nil {
		log.Printf("getShippingRates failed: %v", err)
		return nil, store.Shipment{}, err
	}

	parcels, packages, err := createParcels(c, order.Items, parcelObjs)
	if err != nil {
		log.Printf("getShippingRates failed: %v", err)
		return nil, store.Shipment{}, err
	}

	// create shipment objects and get rates
	parcelIDinput := []string{}
	for _, p := range parcels {
		parcelIDinput = append(parcelIDinput, p.ObjectID)
	}

	shipmentInput := &models.ShipmentInput{
		AddressFrom: from.ObjectID,
		AddressTo:   to.ObjectID,
		Parcels:     parcelIDinput,
	}

	shipment, err := c.CreateShipment(shipmentInput)
	if err != nil {
		log.Printf("getShippingRates failed: %v", err)
		return nil, store.Shipment{}, err
	}

	// return rates & object to store in DB for further actioning
	shipmentDB := createShipmentObject(data, shipment, packages)

	return shipmentDB.Rates, shipmentDB, nil
}

// create shippo address object with customer info
func createShipmentAddress(c *client.Client, data store.Address) (*models.Address, error) {
	ai := &models.AddressInput{
		Name:          data.FirstName + " " + data.LastName,
		Company:       data.Company,
		Street1:       data.AddressLine1,
		Street2:       data.AddressLine2,
		City:          data.City,
		Zip:           data.Zip,
		State:         data.State,
		Country:       data.Country,
		Phone:         data.PhoneNumber,
		Email:         data.Email,
		IsResidential: true,
		Validate:      true,
	}
	// populate other fields if applicable
	addr, err := c.CreateAddress(ai)
	if err != nil {
		log.Printf("createShipmentAddress failed: %v", err)
		return nil, err
	}
	log.Printf("validation result: %v; %v", addr.ValidationResults.IsValid, addr.ValidationResults.Messages)
	if !addr.ValidationResults.IsValid {
		return nil, fmt.Errorf("INVALID_ADDRESS")
	}
	return addr, nil
}

// create shippo address object with business info
func createReturnAddress(c *client.Client) (*models.Address, error) {
	data := store.ReturnAddress
	ai := &models.AddressInput{
		Name:          data.FirstName + " " + data.LastName,
		Company:       data.Company,
		Street1:       data.AddressLine1,
		Street2:       data.AddressLine2,
		City:          data.City,
		Zip:           data.Zip,
		State:         data.State,
		Country:       data.Country,
		Phone:         data.PhoneNumber,
		Email:         data.Email,
		IsResidential: true,
		Validate:      false,
	}
	// populate other fields if applicable
	addr, err := c.CreateAddress(ai)
	if err != nil {
		log.Printf("createShipmentAddress failed: %v", err)
		return nil, err
	}

	return addr, nil
}

// Create parcel(s) for order. Uses greedy algorithm for large multi-parcel orders to fit as many
// objects into the largest parcel as possible (higher price : volume ratio) and fit the remainder in the smallest
// parcel as possible and repeats as necessary for orders requring >2 parcels.
func createParcels(c *client.Client, items []*store.CartItem, parcels []*store.Parcel) ([]*models.Parcel, []store.Package, error) {
	parcelObjs := []*models.Parcel{}
	packages := []store.Package{}
	resVolPct := float32(0.2)

	// split packages with greedy algorithm if total order volume is greater than largest parcel volume
	// sort CartItems by volume greatest to least.
	sortedByVol := sortops.SortCartItemsByUnitVolume(items)
	// create package item for each individual unit in order
	pkgItems := []PkgItem{}
	for _, item := range sortedByVol {
		for j := 0; j < item.Quantity; j++ {
			// create PkgItem for each individual unit
			pd := item.ShippingDimensions
			floats, err := pd.GetFloatsMM()
			if err != nil {
				log.Printf("createParcels failed - get item floats: %v", err)
				return parcelObjs, packages, err
			}
			pi := PkgItem{
				ItemID: item.SizeID,
				Name:   item.Name,
				Length: floats[0],
				Width:  floats[1],
				Height: floats[2],
				Volume: floats[0] * floats[1] * floats[2],
			}
			pkgItems = append(pkgItems, pi)
		}
	}

	// create parcels for order until there is no remaining order volume
	prevRem := 0
	for {
		// get parcel
		parcel, packaged, rem, pack, err := getParcelForVolume(c, parcels, items, pkgItems, resVolPct)
		if err != nil {
			log.Printf("createParcels failed: %v", err)
			return parcelObjs, packages, err
		}

		// create package summary / add packing list to Package
		for _, item := range pack {
			if packaged.Items[item.ItemID] == nil {
				itemSum := &store.PkgItemSummary{
					ItemID:   item.ItemID,
					Name:     item.Name,
					Quantity: 1,
				}
				packaged.Items[item.ItemID] = itemSum
			} else {
				packaged.Items[item.ItemID].Quantity++
			}
		}

		parcelObjs = append(parcelObjs, parcel)
		packages = append(packages, packaged)

		// return if complete order packaged
		if len(rem) == 0 {
			return parcelObjs, packages, nil
		}

		pkgItems = rem
		if len(rem) == prevRem {
			// edge case - no parcel found in list for remaining items
			return []*models.Parcel{}, []store.Package{}, fmt.Errorf("NO_PARCEL_FOUND")
		}
		prevRem = len(rem)
	}
}

// get package dimensions required to fit order
func getDimensions(items []*store.CartItem) (dimensions, error) {
	totalWtLbs := float32(0.0)
	totalVolume := float32(0.0)
	maxLength := float32(0.0)
	maxWidth := float32(0.0)
	maxHeight := float32(0.0)

	// calculate order volume
	for _, item := range items {
		// get volume and weight
		floats, err := item.ShippingDimensions.GetFloatsMM()
		if err != nil {
			log.Printf("getDimensions failed: %v", err)
			return dimensions{}, err
		}
		l, w, h, wt := floats[0], floats[1], floats[2], floats[3]
		volume := l * w * h
		totalWtLbs += (wt * float32(item.Quantity))
		totalVolume += volume

		// get max l, w, h
		if l > maxLength {
			maxLength = l
		}
		if w > maxWidth {
			maxWidth = w
		}
		if h > maxHeight {
			maxHeight = h
		}

		// getWeightLbs()
	}

	dim := dimensions{totalWtLbs, totalVolume, maxLength, maxWidth, maxHeight}

	return dim, nil
}

// get smallest parcel for order volume in cubic mm
func getParcelForVolume(c *client.Client, parcels []*store.Parcel, cartItems []*store.CartItem, pkgItems []PkgItem, resvPct float32) (*models.Parcel, store.Package, []PkgItem, []PkgItem, error) {
	rem := []PkgItem{}
	pack := []PkgItem{}
	sorted := sortops.SortParcelsByVolume(parcels) // sort by volume least to greatest
	pkg := store.Package{}
	pi := &models.ParcelInput{}

	if len(cartItems) == 0 || len(pkgItems) == 0 {
		log.Printf("getParcelForVolume: no items")
		return &models.Parcel{}, store.Package{}, rem, pack, nil
	}

	// get parcel dimension constraints from order in mm
	dimensions, err := getDimensions(cartItems)
	if err != nil {
		log.Printf("getParcelForVolume failed: %v", err)
		return &models.Parcel{}, store.Package{}, rem, pack, err
	}
	volume := dimensions.Volume
	mL, mW, mH := dimensions.MaxLength, dimensions.MaxWidth, dimensions.MaxHeight

	// search for available parcel to fit order volume and product dimension constraints
	for _, p := range sorted {
		floats, err := p.ParcelDimensions.GetFloatsMM()
		if err != nil {
			log.Printf("getParcelForVolume: no items")
			return &models.Parcel{}, store.Package{}, rem, pack, err
		}
		l, w, h := floats[0], floats[1], floats[2]
		parcelVol := l * w * h
		if volume < float32(parcelVol*(1-resvPct)) { // leave extra space for packaging materials
			// order volume ok - verify parcel dimensions fit largest items
			// compare dimensions of largest items to dimensions of parcel in mm
			l, w, h := floats[0], floats[1], floats[2]
			if l < mL || w < mW || h < mH {
				// parcel does not fit largest objects
				continue
			}

			// fill parcel
			rem, pack, err = fillParcel(pkgItems, p, float32(0.1))
			if err != nil {
				log.Printf("getParcelForVolume failed: %v", err)
				return &models.Parcel{}, store.Package{}, rem, pack, err
			}

			// get parcel wt
			pWt, err := p.ParcelDimensions.GetWeightLb()
			if err != nil {
				log.Printf("getParcelForVolume failed: %v", err)
				return &models.Parcel{}, store.Package{}, rem, pack, err
			}
			totalWt := pWt
			inParcel := make(map[string]*store.CartItem)
			for _, item := range cartItems {
				inParcel[item.SizeID] = item
			}
			for _, item := range pack {
				unitWt, err := inParcel[item.ItemID].ShippingDimensions.GetWeightLb()
				if err != nil {
					log.Printf("getParcelForVolume failed: %v", err)
					return &models.Parcel{}, store.Package{}, rem, pack, err
				}
				totalWt += unitWt
			}

			// create store.Package object for DB storage
			pkg = store.Package{
				Carrier:    p.Carrier,
				ParcelID:   p.ParcelID,
				Name:       p.Name,
				Dimensions: p.ParcelDimensions,
				Template:   p.Template,
				Items:      make(map[string]*store.PkgItemSummary),
			}
			// create shippo parcel object
			pi = &models.ParcelInput{
				Length:       p.ParcelDimensions.Length,
				Width:        p.ParcelDimensions.Width,
				Height:       p.ParcelDimensions.Height,
				DistanceUnit: p.ParcelDimensions.DistanceUnit,
				Weight:       fmt.Sprintf("%.2f", totalWt),
				MassUnit:     p.ParcelDimensions.MassUnit,
			}
			if len(rem) > 0 {
				// try next largest parcel to attempt to fit whole order in one package
				continue
			}
			break

		}
	}

	// shippo parcel object
	parcel, err := c.CreateParcel(pi)
	if err != nil {
		log.Printf("getParcelForVolume failed: %v", err)
		return &models.Parcel{}, store.Package{}, rem, pack, err
	}

	return parcel, pkg, rem, pack, nil
}

// fillParcel fills the selected Parcel with the Order's items and
// returns a list of any remaining items, the parcel's packing list, and an error value.
func fillParcel(items []PkgItem, parcel *store.Parcel, resv float32) ([]PkgItem, []PkgItem, error) {
	rem := []PkgItem{}
	pack := []PkgItem{}

	if len(items) == 0 {
		log.Println("fill parcel - empty item list")
		return rem, pack, nil
	}

	d := parcel.ParcelDimensions
	floats, err := d.GetFloatsMM()
	if err != nil {
		log.Printf("fillParcel failed - get parcel floats: %v", err)
		return rem, pack, err
	}
	mL, mW, mH := floats[0], floats[1], floats[2]

	box := &Box{
		Length:  mL,
		Width:   mW,
		Height:  mH,
		Volume:  mL * mW * mH,
		ResvPct: resv,
	}

	// get remaining and packaged items from greedy algorithm
	rem, pack = addToBox(items, box)

	return rem, pack, nil
}

// addToBox recursively adds a list of PkgItems to a Root Box (representing a Parcel object),
// until there are no remaining items, or there is an insufficient amount of space
// left in the Root Box for the remaining items.
// The remaining items and the list of packed items are returned to the caller.
func addToBox(items []PkgItem, box *Box) ([]PkgItem, []PkgItem) {
	rem := []PkgItem{}  // remaining items
	pack := []PkgItem{} // packing list

	if len(items) == 0 {
		// edge case
		return rem, pack
	}
	if box.Length == 0.0 || box.Width == 0.0 || box.Height == 0.0 {
		// no remaining space along either axis
		return items, pack
	}

	// add current item to current box; create child nodes
	err := box.Add(items[0])
	if err != nil {
		// next largest does not fit; add to remaining
		rem = append(rem, items[0])
		if len(items) > 1 {
			// fill with remaining items if amy
			r, p := addToBox(items[1:], box)
			rem = append(rem, r...)
			pack = append(pack, p...)
			return rem, pack
		} else {
			// no remaining items
			return rem, pack
		}
	} else {
		// item fits in box - add item to packing list
		pack = append(pack, items[0])
	}
	if len(items) > 1 {
		// fill Length branch/Box with remaining items
		res, pk := addToBox(items[1:], box.NodeL)
		pack = append(pack, pk...)
		if len(res) != 0 {
			// Length branch does not contain sufficient space to fill total remaining items
			// fill Width branch/Box with remaining items from Length branch
			res, pk = addToBox(res, box.NodeW)
			pack = append(pack, pk...)
		} else {
			// total of remainder of items packed into Length Branch
			return rem, pack
		}
		if len(res) != 0 {
			// Width branch does not contain sufficient space to fill remainder of
			// items from Length Branch
			// fill Height branch/Box with remaining items from Width branch
			res, pk = addToBox(res, box.NodeH)
		} else {
			// total of remainder of items packed into Width Branch
			return rem, pack
		}

		// add any remaining and packed items
		rem = append(rem, res...)  // remainder of items that could not fit in Height branch (if any)
		pack = append(pack, pk...) // items packed in Height branch (if any)
	}
	return rem, pack
}

// create store.Shipment object for order fullfillment
func createShipmentObject(user customerInfo, s *models.Shipment, pkgs []store.Package) store.Shipment {
	addr := store.Address{
		FirstName:    s.AddressTo.Name,
		Company:      s.AddressTo.Company,
		AddressLine1: s.AddressTo.Street1,
		AddressLine2: s.AddressTo.Street2,
		City:         s.AddressTo.City,
		State:        s.AddressTo.State,
		Country:      s.AddressTo.Country,
		Zip:          s.AddressTo.Zip,
		PhoneNumber:  s.AddressTo.Phone,
		Email:        s.AddressTo.Email,
	}

	rates := []store.RateSummary{}
	for _, rate := range s.Rates {
		if rate.Provider != "USPS" {
			continue
		}
		sl := store.ServiceLevel{
			Name:  rate.ServiceLevel.Name,
			Token: rate.ServiceLevel.Token,
			Terms: rate.ServiceLevel.Terms,
		}
		p, _ := strconv.ParseFloat(rate.AmountLocal, 32)
		rs := store.RateSummary{
			Price:        rate.AmountLocal,
			PriceFloat:   float32(p),
			Currency:     rate.Currency,
			Provider:     rate.Provider,
			Days:         rate.Days,
			ServiceLevel: sl,
		}
		rates = append(rates, rs)
	}

	shipment := store.Shipment{
		UserID:      user.UserID,
		OrderID:     user.OrderID,
		Status:      s.Status,
		AddressTo:   addr,
		AddressFrom: store.ReturnAddress,
		Packages:    pkgs,
		Rates:       rates,
	}

	return shipment
}

func main() {
	httpops.RegisterRoutes(route, RootHandler)
	log.Fatal(gateway.ListenAndServe(":3000", nil))
}
