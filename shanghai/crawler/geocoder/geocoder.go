package geocoder

import (
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type Address struct {
	Address   string
	Longitude float64
	Latitude  float64
}

type Geocoder struct {
	api   GeocoderAPI
	cache *GeocodeCache
}

func (g Geocoder) Geocode(addr string) (*Address, error) {
	//	先检查缓存是否已存在该地址的解析
	if g.cache != nil {
		lon, lat, err := g.cache.Get(addr)
		if err != nil {
			// log.Warnf("Geocode(%s): 读取失败：%s", addr, err)
		} else {
			return &Address{Address: addr, Longitude: lon, Latitude: lat}, nil
		}
	}

	//	缓存中不存在该地址，因此需要调用API解析
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

	//	将坐标信息保存于缓存
	if g.cache != nil {
		err := g.cache.Put(a.Address, a.Longitude, a.Latitude)
		if err != nil {
			log.Errorf("Geocode(%q): 写入失败：%s", addr, err)
		}
	}
	return a, nil
}

type GeocoderAPI interface {
	GetURL(addr string) *url.URL
	Parse(body []byte) (*Address, error)
}
