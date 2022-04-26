package geocoder

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
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

func NewGeocoderAMAP(key string) Geocoder {
	return Geocoder{api: GeocoderAPIAmap{key: key}}
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
	//	返回
	return &gi, nil
}
