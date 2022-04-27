package geocoder

import (
	"encoding/json"
	"fmt"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/suifengtec/gocoord"
)

// https://lbsyun.baidu.com/index.php?title=webapi/guide/changeposition
// https://lbsyun.baidu.com/index.php?title=coordinate

// 天地图
//	[Request]
//		https://api.map.baidu.com/geocoding/v3/?address=北京市海淀区上地十街10号&output=json&ak=您的ak&callback=showLocation
//	[Response]
// {
// 	"status": 0,
// 	"result": {
// 	  "location": { "lng": 121.4638275031133, "lat": 31.263190279921234 },
// 	  "precise": 1,
// 	  "confidence": 80,
// 	  "comprehension": 100,
// 	  "level": "门址"
// 	}
// }

type GeocoderAPIBaiduResponse struct {
	Status int
	Result struct {
		Location struct {
			Lng float64
			Lat float64
		}
		Precise       int
		Confidence    int
		Comprehension int
		Level         string
	}
}

type GeocoderAPIBaidu struct {
	key string
}

func NewGeocoderBaidu(key, cachedir string) Geocoder {
	var cache *GeocodeCache
	if len(cachedir) > 0 {
		var err error
		cache, err = NewGeocodeCache(cachedir)
		if err != nil {
			log.Errorf("NewGeocoderBaidu(): 无法建立缓存[%s]：%s", cachedir, err)
		}
	}
	return Geocoder{api: GeocoderAPIBaidu{key: key}, cache: cache}
}

func (a GeocoderAPIBaidu) GetURL(addr string) *url.URL {
	u, err := url.Parse("https://api.map.baidu.com/geocoding/v3/")
	if err != nil {
		return nil
	}
	params := u.Query()
	params.Add("address", addr)
	params.Add("ret_coordtype", "gcj02ll") // gcj02ll: 国测局坐标; bd09mc: 百度墨卡托坐标; bd09ll: 百度经纬度坐标
	params.Add("output", "json")
	params.Add("ak", a.key)
	u.RawQuery = params.Encode()

	return u
}

func (a GeocoderAPIBaidu) Parse(body []byte) (*Address, error) {
	gi := Address{}
	//	解析 JSON
	r := GeocoderAPIBaiduResponse{}

	var err error
	if err = json.Unmarshal(body, &r); err != nil {
		return &gi, err
	}
	//	处理API返回结果
	resp := GeocoderAPIBaiduResponse{}
	json.Unmarshal(body, &resp)
	if resp.Status != 0 {
		return nil, fmt.Errorf("GeocoderAPIBaidu.Parse(): [%d]", resp.Status)
	}
	//	坐标转换：GCJ02 => WGS84
	l2 := gocoord.GCJ02ToWGS84(gocoord.Position{Lon: resp.Result.Location.Lng, Lat: resp.Result.Location.Lat})
	// fmt.Printf("GCJ02 (%.6f, %.6f) => WGS84 (%.6f, %6f)\n",
	// 	resp.Result.Location.Lat, resp.Result.Location.Lng,
	// 	l2.Lat, l2.Lon,
	// )
	//	返回
	return &Address{
		Longitude: l2.Lon,
		Latitude:  l2.Lat,
	}, nil
}
