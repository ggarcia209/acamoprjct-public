package main

import (
	"log"
	"os"
	"testing"

	"github.com/ggarcia209/acamoprjct/service/store-api/store"
	"github.com/ggarcia209/acamoprjct/service/util/dbops"
	"github.com/ggarcia209/acamoprjct/service/util/shipops"
	"github.com/ggarcia209/acamoprjct/service/util/sortops"
)

func getTestToken() (string, error) {
	testToken, err := shipops.GetToken("./shippo_test_tk.txt")
	if err != nil {
		log.Printf("getTestToken failed: %v", err)
		return "", err
	}
	return testToken, nil
}

func TestCreateReturnAddress(t *testing.T) {
	token, err := getTestToken()
	if err != nil {
		t.Errorf("FAIL - get token: %v", err)
	}

	client := shipops.InitClient(token)
	addr, err := createReturnAddress(client)
	if err != nil {
		t.Errorf("FAIL: %v", err)
	}
	t.Logf("%v", addr)
}

func TestAddToBox(t *testing.T) {
	var tests = []struct {
		items   []PkgItem
		box     *Box
		wantIds []string
		wantErr error
	}{
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			box: &Box{
				Length:  12.0,
				Width:   12.0,
				Height:  6.0,
				Volume:  864.0,
				ResvPct: 0.00,
			},
			wantIds: []string{"001", "002", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			box: &Box{
				Length:  12.0,
				Width:   12.0,
				Height:  6.0,
				Volume:  864.0,
				ResvPct: 0.20,
			},
			wantIds: []string{"001", "002", "003", "003", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			box: &Box{
				Length:  6.0,
				Width:   6.0,
				Height:  6.0,
				Volume:  864.0,
				ResvPct: 0.20,
			},
			wantIds: []string{"002", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 2.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			box: &Box{
				Length:  6.0,
				Width:   6.0,
				Height:  6.0,
				Volume:  216.0,
				ResvPct: 0.00,
			},
			wantIds: []string{"002", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "002", Name: "Item 2", Length: 6.0, Width: 6.0, Height: 2.0, Volume: 72.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			box: &Box{
				Length:  6.0,
				Width:   6.0,
				Height:  6.0,
				Volume:  216.0,
				ResvPct: 0.00,
			},
			wantIds: []string{"002", "003", "003", "003", "003", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "002", Name: "Item 2", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
			},
			box: &Box{
				Length:  6.0,
				Width:   6.0,
				Height:  6.0,
				Volume:  216.0,
				ResvPct: 0.00,
			},
			wantIds: []string{},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "002", Name: "Item 2", Length: 6.0, Width: 6.0, Height: 2.0, Volume: 72.0},
			},
			box: &Box{
				Length:  6.0,
				Width:   6.0,
				Height:  6.0,
				Volume:  216.0,
				ResvPct: 0.00,
			},
			wantIds: []string{"002"},
			wantErr: nil,
		},
		{
			items: []PkgItem{},
			box: &Box{
				Length:  6.0,
				Width:   6.0,
				Height:  6.0,
				Volume:  216.0,
				ResvPct: 0.00,
			},
			wantIds: []string{},
			wantErr: nil,
		},
	}

	for _, test := range tests {
		t.Logf("*** TEST ****")
		rem, pack := addToBox(test.items, test.box)
		t.Logf("rem: %v", rem)
		t.Logf("pack: %v", pack)
		t.Log("")
		t.Log("")
		if len(pack) != len(test.wantIds) {
			t.Errorf("FAIL - data: %d; want: %d", len(pack), len(test.wantIds))
		}
		for i, item := range pack {
			if item.ItemID != test.wantIds[i] {
				t.Errorf("FAIL - data: %s; want: %s", item.ItemID, test.wantIds[i])
			}
		}
	}
}

func TestFillParcel(t *testing.T) {
	var tests = []struct {
		parcel  *store.Parcel
		items   []PkgItem
		resv    float32
		wantIds []string
		wantErr error
	}{
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_largeflatratebox",
				Name:     "USPS Large Flat Rage Box",
				ParcelDimensions: store.Dimensions{
					Length: "12.0",
					Width:  "12.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 864.0,
				},
			},
			resv:    0.2,
			wantIds: []string{"001", "002", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_largeflatratebox",
				Name:     "USPS Large Flat Rage Box",
				ParcelDimensions: store.Dimensions{
					Length: "12.0",
					Width:  "12.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 864.0,
				},
			},
			resv:    0.2,
			wantIds: []string{"001", "002", "003", "003", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_squarebox",
				Name:     "USPS Square Box",
				ParcelDimensions: store.Dimensions{
					Length: "6.0",
					Width:  "6.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 216.0,
				},
			},
			resv:    0.2,
			wantIds: []string{"002", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "001", Name: "Item 1", Length: 8.0, Width: 8.0, Height: 2.0, Volume: 192.0},
				PkgItem{ItemID: "002", Name: "Item 2", Length: 5.0, Width: 5.0, Height: 2.0, Volume: 50.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_squarebox",
				Name:     "USPS Square Box",
				ParcelDimensions: store.Dimensions{
					Length: "6.0",
					Width:  "6.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 216.0,
				},
			},
			resv:    0.0,
			wantIds: []string{"002", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "002", Name: "Item 2", Length: 6.0, Width: 6.0, Height: 2.0, Volume: 72.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
				PkgItem{ItemID: "003", Name: "Item 3", Length: 3.0, Width: 3.0, Height: 2.0, Volume: 18.0},
			},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_squarebox",
				Name:     "USPS Square Box",
				ParcelDimensions: store.Dimensions{
					Length: "6.0",
					Width:  "6.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 216.0,
				},
			},
			resv:    0.0,
			wantIds: []string{"002", "003", "003", "003", "003", "003", "003", "003", "003"},
			wantErr: nil,
		},
		{
			items: []PkgItem{
				PkgItem{ItemID: "002", Name: "Item 2", Length: 8.0, Width: 8.0, Height: 3.0, Volume: 192.0},
			},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_squarebox",
				Name:     "USPS Square Box",
				ParcelDimensions: store.Dimensions{
					Length: "6.0",
					Width:  "6.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 216.0,
				},
			},
			resv:    0.0,
			wantIds: []string{},
			wantErr: nil,
		},
		{ // edge case - single item fills 100% of volume
			items: []PkgItem{
				PkgItem{ItemID: "002", Name: "Item 2", Length: 6.0, Width: 6.0, Height: 2.0, Volume: 72.0},
			},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_squarebox",
				Name:     "USPS Square Box",
				ParcelDimensions: store.Dimensions{
					Length: "6.0",
					Width:  "6.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 216.0,
				},
			},
			resv:    0.0,
			wantIds: []string{"002"},
			wantErr: nil,
		},
		{ // edge case - empty list
			items: []PkgItem{},
			parcel: &store.Parcel{
				Carrier:  "usps",
				ParcelID: "usps_squarebox",
				Name:     "USPS Square Box",
				ParcelDimensions: store.Dimensions{
					Length: "6.0",
					Width:  "6.0",
					Height: "6.0",
					Weight: "1.0",
					Volume: 216.0,
				},
			},
			resv:    0.0,
			wantIds: []string{},
			wantErr: nil,
		},
	}
	for _, test := range tests {
		rem, pack, err := fillParcel(test.items, test.parcel, test.resv)
		if err != test.wantErr {
			t.Errorf("FAIL: %v; want: %v", err, test.wantErr)
		}
		if len(pack) != len(test.wantIds) {
			t.Errorf("FAIL - data: %d; want: %d", len(pack), len(test.wantIds))
			t.Logf("items: %v", test.items)
			t.Logf("rem: %v", rem)
			t.Logf("pack: %v", pack)
			return
		}
		for i, item := range pack {
			if item.ItemID != test.wantIds[i] {
				t.Errorf("FAIL - data: %s; want: %s", item.ItemID, test.wantIds[i])
			}
		}
		t.Logf("rem: %v", rem)
		t.Logf("pack: %v", pack)
	}
}

func TestGetDimensions(t *testing.T) {
	var tests = []struct {
		items  []*store.CartItem
		length float32
		width  float32
		height float32
		weight float32
		volume float32
	}{
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "8.0", Width: "8.0", Height: "3.0", Weight: "1.0", Volume: 192.00},
				},
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			length: 8.0,
			width:  8.0,
			height: 3.0,
			weight: 2.0,
			volume: 260,
		},
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "8.0", Width: "8.0", Height: "3.0", Weight: "1.0", Volume: 192.00},
				},
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			length: 8.0,
			width:  8.0,
			height: 3.0,
			weight: 3.5,
			volume: 314,
		},
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
			},
			length: 5.0,
			width:  5.0,
			height: 2.0,
			weight: 0.5,
			volume: 50.0,
		},
		{
			items:  []*store.CartItem{},
			length: 0.0,
			width:  0.0,
			height: 0.0,
			weight: 0.0,
			volume: 0.0,
		},
	}
	for _, test := range tests {
		d, err := getDimensions(test.items)
		if err != nil {
			t.Errorf("FAIL: %v", err)
			return
		}
		if d.MaxLength != test.length {
			t.Errorf("FAIL - length: %f; want: %f", d.MaxLength, test.length)
		}
		if d.MaxWidth != test.width {
			t.Errorf("FAIL - width: %f; want: %f", d.MaxWidth, test.width)
		}
		if d.MaxHeight != test.height {
			t.Errorf("FAIL - height: %f; want: %f", d.MaxHeight, test.height)
		}
		if d.Weight != test.weight {
			t.Errorf("FAIL - weight: %f; want: %f", d.Weight, test.weight)
		}
		if d.Volume != test.volume {
			t.Errorf("FAIL - volume: %f; want: %f", d.Volume, test.volume)
		}
	}
}

func TestGetParcelForVolume(t *testing.T) {
	var tests = []struct {
		items   []*store.CartItem
		resv    float32
		wantErr error
	}{
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "8.0", Width: "8.0", Height: "3.0", Weight: "1.0", Volume: 192.00},
				},
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "003",
					SizeID:             "003-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			resv:    0.0,
			wantErr: nil,
		},
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "8.0", Width: "8.0", Height: "3.0", Weight: "1.0", Volume: 192.00},
				},
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "003",
					SizeID:             "003-OS",
					Quantity:           4,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			resv:    0.2,
			wantErr: nil,
		},
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "001",
					SizeID:             "001-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "8.0", Width: "8.0", Height: "3.0", Weight: "1.0", Volume: 192.00},
				},
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "003",
					SizeID:             "003-OS",
					Quantity:           4,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			resv:    0.0,
			wantErr: nil,
		},
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "6.0", Width: "6.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "003",
					SizeID:             "003-OS",
					Quantity:           8,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			resv:    0.0,
			wantErr: nil,
		},
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "6.0", Width: "6.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "003",
					SizeID:             "003-OS",
					Quantity:           8,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			resv:    0.2,
			wantErr: nil,
		},
		{ // 1 large flate rate + 1 square
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           4,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "5.0", Weight: "0.5", Volume: 125.00},
				},
				&store.CartItem{
					ItemID:             "003",
					SizeID:             "003-OS",
					Quantity:           10,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
				},
			},
			resv:    0.0,
			wantErr: nil,
		},
		{
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
				},
			},
			resv:    0.0,
			wantErr: nil,
		},
		{
			items:   []*store.CartItem{},
			resv:    0.0,
			wantErr: nil,
		},
	}

	token, err := getTestToken()
	if err != nil {
		t.Errorf("FAIL - get token: %v", err)
	}
	client := shipops.InitClient(token)

	os.Setenv(dbops.EnvarParcelsTable, "acamoprjct-parcels-dev")
	os.Setenv(dbops.EnvarStoreItemsIndexTable, "acamoprjct-store-items-index-dev")
	table1 := dbops.NewTable(dbops.ParcelsTable(), dbops.ParcelsPK, "")
	table2 := dbops.NewTable(dbops.StoreItemsIndexTable(), dbops.StoreItemsIndexPK, "")
	tables := []dbops.Table{table1, table2}
	dbInfo := dbops.InitDB(tables)

	indexKey := "parcels-usps"
	index, err := dbops.GetStoreItemIndex(dbInfo, indexKey)
	if err != nil {
		t.Errorf("FAIL - index: %v", err)
		return
	}
	t.Logf("index: %v", index.ItemIDs)
	parcels, err := dbops.BatchGetParcels(dbInfo, index.ItemIDs)
	if err != nil {
		t.Errorf("FAIL - get parcels: %v", err)
		return
	}
	t.Logf("%v", parcels)

	for _, test := range tests {
		// split packages with greedy algorithm if total order volume is greater than largest parcel volume
		sortedByVol := sortops.SortCartItemsByUnitVolume(test.items)
		// create package item for each individual unit in order
		pkgItems := []PkgItem{}
		for _, item := range sortedByVol {
			for j := 0; j < item.Quantity; j++ {
				pd := item.ShippingDimensions
				floats, err := pd.GetFloats()
				if err != nil {
					t.Errorf("FAIL - get parcels: %v", err)
					return
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
		parcel, pkg, rem, pack, err := getParcelForVolume(client, parcels, sortedByVol, pkgItems, test.resv)
		if err != test.wantErr {
			t.Errorf("FAIL: %v; want: %v", err, test.wantErr)
			return
		}
		t.Logf("parcel: %v", parcel)
		t.Logf("package: %v", pkg)
		t.Logf("rem: %v", rem)
		t.Logf("pack: %v", pack)
	}
}

func TestCreateParcels(t *testing.T) {
	var tests = []struct {
		items   []*store.CartItem
		wantErr error
	}{
		{ // 1 square box
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "002",
					SizeID:             "002-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", DistanceUnit: "in", Weight: "0.5", MassUnit: "lb", Volume: 50.00},
				},
				&store.CartItem{
					ItemID:             "003",
					SizeID:             "003-OS",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", DistanceUnit: "in", Weight: "0.5", MassUnit: "lb", Volume: 18.00},
				},
			},
			wantErr: nil,
		},
		{ // 1 square box
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "004",
					SizeID:             "004-13x13",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "5", Width: "5", Height: "5", DistanceUnit: "in", Weight: "1", MassUnit: "lb", Volume: 125.00, VolumeUnit: "in"},
				},
			},
			wantErr: nil,
		},
		{ // 1 square box
			items: []*store.CartItem{
				&store.CartItem{
					ItemID:             "004",
					SizeID:             "004-13x13",
					Quantity:           1,
					ShippingDimensions: store.Dimensions{Length: "3", Width: "3", Height: "3", DistanceUnit: "in", Weight: "1", MassUnit: "lb", Volume: 27.00, VolumeUnit: "in"},
				},
			},
			wantErr: nil,
		}, /*
			{ // 1 large flat rate
				items: []*store.CartItem{
					&store.CartItem{
						ItemID:             "001",
						SizeID:             "001-OS",
						Quantity:           1,
						ShippingDimensions: store.Dimensions{Length: "8.0", Width: "8.0", Height: "3.0", Weight: "1.0", Volume: 192.00},
					},
					&store.CartItem{
						ItemID:             "002",
						SizeID:             "002-OS",
						Quantity:           1,
						ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
					},
					&store.CartItem{
						ItemID:             "003",
						SizeID:             "003-OS",
						Quantity:           4,
						ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
					},
				},
				wantErr: nil,
			},
			{ // 2 large flat rate
				items: []*store.CartItem{
					&store.CartItem{
						ItemID:             "001",
						SizeID:             "001-OS",
						Quantity:           3,
						ShippingDimensions: store.Dimensions{Length: "8.0", Width: "8.0", Height: "3.0", Weight: "1.0", Volume: 192.00},
					},
					&store.CartItem{
						ItemID:             "002",
						SizeID:             "002-OS",
						Quantity:           1,
						ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
					},
					&store.CartItem{
						ItemID:             "003",
						SizeID:             "003-OS",
						Quantity:           6,
						ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
					},
				},
				wantErr: nil,
			},
			{ // 1 large flat rate + 1 square
				items: []*store.CartItem{
					&store.CartItem{
						ItemID:             "002",
						SizeID:             "002-OS",
						Quantity:           4,
						ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "5.0", Weight: "0.5", Volume: 125.00},
					},
					&store.CartItem{
						ItemID:             "003",
						SizeID:             "003-OS",
						Quantity:           10,
						ShippingDimensions: store.Dimensions{Length: "3.0", Width: "3.0", Height: "2.0", Weight: "0.5", Volume: 18.00},
					},
				},
				wantErr: nil,
			},
			{ // 1 square - single item
				items: []*store.CartItem{
					&store.CartItem{
						ItemID:             "002",
						SizeID:             "002-OS",
						Quantity:           1,
						ShippingDimensions: store.Dimensions{Length: "5.0", Width: "5.0", Height: "2.0", Weight: "0.5", Volume: 50.00},
					},
				},
				wantErr: nil,
			},
			{ // no parcel found for size
				items:   []*store.CartItem{},
				wantErr: nil,
			}, */
	}

	token, err := getTestToken()
	if err != nil {
		t.Errorf("FAIL - get token: %v", err)
	}
	client := shipops.InitClient(token)

	os.Setenv(dbops.EnvarParcelsTable, "acamoprjct-parcels-dev")
	os.Setenv(dbops.EnvarStoreItemsIndexTable, "acamoprjct-store-items-index-dev")
	os.Setenv(dbops.EnvarOrdersTable, "acamoprjct-orders-dev")
	table1 := dbops.NewTable(dbops.ParcelsTable(), dbops.ParcelsPK, "")
	table2 := dbops.NewTable(dbops.StoreItemsIndexTable(), dbops.StoreItemsIndexPK, "")
	tables := []dbops.Table{table1, table2}
	dbInfo := dbops.InitDB(tables)

	indexKey := "parcels-" + store.CarriersUsps
	parcelIDs, err := dbops.GetStoreItemIndex(dbInfo, indexKey) // fixed to USPS
	if err != nil {
		t.Errorf("FAIL - get parcel index: %v", err)
	}

	parcels, err := dbops.BatchGetParcels(dbInfo, parcelIDs.ItemIDs)
	if err != nil {
		t.Errorf("FAIL - get parcels: %v", err)
	}

	t.Logf("parcelIDs: %v", parcelIDs.ItemIDs)

	for _, test := range tests {
		t.Log("*** TEST ***")
		parcels, packages, err := createParcels(client, test.items, parcels)
		if err != test.wantErr {
			t.Errorf("FAIL: %v; want: %v", err, test.wantErr)
		}
		for _, p := range parcels {
			t.Logf("parcel: %v", p)
		}
		for _, p := range packages {
			t.Logf("package: %v", p)
		}
		t.Log("----------")
		t.Log("")
	}

}

func TestGetShippingRates(t *testing.T) {
	var tests = []struct {
		info    customerInfo
		wantErr error
	}{
		/* {info: customerInfo{UserID: "659cbc67021545a8e14e1a00cf0d48d2", OrderID: "659cbc67021545a8e14e1a00cf0d48d2-1"}, wantErr: nil}, // ok
		{info: customerInfo{UserID: "659cbc67021545a8e14e1a00cf0d48d2", OrderID: "659cbc67021545a8e14e1a00cf0d48d2-2"}, wantErr: nil}, // ok */
		{info: customerInfo{UserID: "659cbc67021545a8e14e1a00cf0d48d2", OrderID: "659cbc67021545a8e14e1a00cf0d48d2-8"}, wantErr: nil},  // ok
		{info: customerInfo{UserID: "659cbc67021545a8e14e1a00cf0d48d2", OrderID: "659cbc67021545a8e14e1a00cf0d48d2-10"}, wantErr: nil}, // ok
		/* {info: customerInfo{UserID: "659cbc67021545a8e14e1a00cf0d48d2", OrderID: "659cbc67021545a8e14e1a00cf0d48d2-9"}, wantErr: nil}, // non-existent order
		{info: customerInfo{UserID: "fakeCusty1", OrderID: "fakeCusty1-3"}, wantErr: nil},        */ // non-existent customer                                                              // no pk or sk
	}

	token, err := getTestToken()
	if err != nil {
		t.Errorf("FAIL - get token: %v", err)
	}
	client := shipops.InitClient(token)

	os.Setenv(dbops.EnvarParcelsTable, "acamoprjct-parcels-dev")
	os.Setenv(dbops.EnvarStoreItemsIndexTable, "acamoprjct-store-items-index-dev")
	os.Setenv(dbops.EnvarOrdersTable, "acamoprjct-orders-dev")
	table1 := dbops.NewTable(dbops.ParcelsTable(), dbops.ParcelsPK, "")
	table2 := dbops.NewTable(dbops.StoreItemsIndexTable(), dbops.StoreItemsIndexPK, "")
	table3 := dbops.NewTable(dbops.OrdersTable(), dbops.OrdersPK, dbops.OrdersSK)
	tables := []dbops.Table{table1, table2, table3}
	dbInfo := dbops.InitDB(tables)

	for _, test := range tests {
		// get order
		order, err := dbops.GetOrder(dbInfo, test.info.UserID, test.info.OrderID)
		if err != nil {
			t.Errorf("FAIL - get order: %v", err)
			continue
		}
		for _, item := range order.Items {
			t.Logf("order items: %v", item)
		}

		t.Logf("parcels table: %v", dbInfo.Tables[dbops.ParcelsTable()])

		rates, shipment, err := getShippingRates(dbInfo, client, test.info, order)
		if err != test.wantErr {
			t.Errorf("FAIL: %v; want: %v", err, test.wantErr)
			continue
		}
		t.Logf("rates: %v", rates)
		t.Logf("shipment: %v", shipment)
	}
}
