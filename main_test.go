package main

import (
	"os"
	"testing"
)

func TestParseProductPageHTML(t *testing.T) {
	f, err := os.Open("product_page.html")
	if err != nil {
		t.Fatal(err)
	}

	p, err := ParseProductPage(f)
	if err != nil {
		t.Fatal(err)
	}

	if *p != (ProductInfo{
		Title:   "The Fat-Loss Plan: 100 Quick and Easy Recipes with Workouts",
		Price:   "£8.49",
		Image:   "https://images-na.ssl-images-amazon.com/images/I/51IsTylYiPL._SX382_BO1,204,203,200_.jpg",
		InStock: true,
	}) {
		t.Fatal(p)
	}
}
