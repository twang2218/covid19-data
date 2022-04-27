package geocoder

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/suifengtec/gocoord"
)

// 高德地图
//
/// https://lbs.amap.com/api/webservice/guide/api/georegeo
//
//	[Request]
//		https://restapi.amap.com/v3/geocode/geo?key={key}&address=上海市徐汇区斜土路2365弄
//	[Response]
// {
// 	"status": "1",
// 	"info": "OK",
// 	"infocode": "10000",
// 	"count": "1",
// 	"geocodes": [
// 	  {
// 		"formatted_address": "上海市徐汇区斜土路2365弄",
// 		"country": "中国",
// 		"province": "上海市",
// 		"citycode": "021",
// 		"city": "上海市",
// 		"district": "徐汇区",
// 		"township": [],
// 		"neighborhood": { "name": [], "type": [] },
// 		"building": { "name": [], "type": [] },
// 		"adcode": "310104",
// 		"street": "斜土路2365弄",
// 		"number": [],
// 		"location": "121.448647,31.192673",
// 		"level": "道路"
// 	  }
// 	]
//   }

type GeocoderAPIAmapResponseItem struct {
	Formatted_Address string
	Country           string
	Province          string
	CityCode          string
	City              string
	District          string
	ADCode            string
	Location          string
	Level             string
}

type GeocoderAPIAmapResponse struct {
	Status   string
	Info     string
	Count    string
	Geocodes []GeocoderAPIAmapResponseItem
}

type GeocoderAPIAmap struct {
	key string
}

func NewGeocoderAMAP(key, cachedir string) Geocoder {
	var cache *GeocodeCache
	if len(cachedir) > 0 {
		var err error
		cache, err = NewGeocodeCache(cachedir)
		if err != nil {
			log.Errorf("NewGeocoderBaidu(): 无法建立缓存[%s]：%s", cachedir, err)
		}
	}
	return Geocoder{api: GeocoderAPIAmap{key: key}, cache: cache}
}

func (a GeocoderAPIAmap) GetURL(addr string) *url.URL {
	u, err := url.Parse("https://restapi.amap.com/v3/geocode/geo")
	if err != nil {
		return nil
	}
	params := u.Query()
	params.Add("key", a.key)
	params.Add("address", addr)
	u.RawQuery = params.Encode()

	return u
}

func (a GeocoderAPIAmap) Parse(body []byte) (*Address, error) {
	gi := Address{}
	//	Parse JSON
	r := GeocoderAPIAmapResponse{}

	var err error
	if err = json.Unmarshal(body, &r); err != nil {
		return &gi, err
	}
	if r.Status != "1" {
		return nil, fmt.Errorf("GeocoderAPIAmap.Parse(): [%s] %s", r.Status, r.Info)
	}
	//	Transform
	rr := r.Geocodes[0]
	gi.Address = rr.Formatted_Address
	///	解析经纬度
	loc := strings.Split(rr.Location, ",")
	if gi.Longitude, err = strconv.ParseFloat(loc[0], 64); err != nil {
		return &gi, fmt.Errorf("无法解析经纬度：%v: %s", rr, err)
	}
	if gi.Latitude, err = strconv.ParseFloat(loc[1], 64); err != nil {
		return &gi, fmt.Errorf("无法解析经纬度：%v: %s", rr, err)
	}
	//	坐标转换：GCJ02 => WGS84
	l2 := gocoord.GCJ02ToWGS84(gocoord.Position{Lon: gi.Longitude, Lat: gi.Latitude})
	// fmt.Printf("GCJ02 (%.6f, %.6f) => WGS84 (%.6f, %6f)\n",
	// 	gi.Latitude, gi.Longitude,
	// 	l2.Lat, l2.Lon,
	// )
	gi.Latitude = l2.Lat
	gi.Longitude = l2.Lon
	//	返回
	return &gi, nil
}
