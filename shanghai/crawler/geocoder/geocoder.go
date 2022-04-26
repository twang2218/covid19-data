package geocoder

import (
	"io"
	"net/http"
	"net/url"
)

type Address struct {
	Address   string
	Longitude float64
	Latitude  float64
}

type Geocoder struct {
	api GeocoderAPI
}

func (g Geocoder) Geocode(addr string) (*Address, error) {
	u := g.api.GetURL(addr)
	//	Request
	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	//	Load Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return g.api.Parse(body)
}

type GeocoderAPI interface {
	GetURL(addr string) *url.URL
	Parse(body []byte) (*Address, error)
}
