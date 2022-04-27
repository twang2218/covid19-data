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
	a, err := g.api.Parse(body)
	if err != nil {
		return a, err
	}
	if len(a.Address) == 0 {
		a.Address = addr
	}
	return a, nil
}

type GeocoderAPI interface {
	GetURL(addr string) *url.URL
	Parse(body []byte) (*Address, error)
}
