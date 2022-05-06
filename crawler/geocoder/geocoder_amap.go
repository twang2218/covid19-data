package geocoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
type GeocoderAPIAmap struct {
	key string
}

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

type GeocoderAPIAmapBatchRequestItem struct {
	Url string `json:"url"`
}

type GeocoderAPIAmapBatchRequest struct {
	Ops []GeocoderAPIAmapBatchRequestItem `json:"ops"`
}

type GeocoderAPIAmapBatchResponse []struct {
	Status int
	Body   GeocoderAPIAmapResponse
}

func NewGeocoderAMAP(key, cachedir string) Geocoder {
	var cache *GeocodeCache
	if len(cachedir) > 0 {
		var err error
		cache, err = NewGeocodeCache(cachedir)
		if err != nil {
			log.Errorf("NewGeocoderAMAP(): 无法建立缓存[%s]：%s", cachedir, err)
		}
	}
	return Geocoder{api: GeocoderAPIAmap{key: key}, cache: cache}
}

func (a GeocoderAPIAmap) Name() string {
	return "高德地图API"
}

func (a GeocoderAPIAmap) Request(addr string) (*Address, error) {
	var err error

	u, err := url.Parse("https://restapi.amap.com/v3/geocode/geo")
	if err != nil {
		return nil, err
	}
	params := u.Query()
	params.Add("key", a.key)
	params.Add("address", addr)
	u.RawQuery = params.Encode()

	resp, err := http.Get(u.String())
	if err != nil {
		return nil, err
	}
	//	Load Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//	Parse JSON
	r := GeocoderAPIAmapResponse{}
	if err = json.Unmarshal(body, &r); err != nil {
		return nil, err
	}

	if r.Status != "1" {
		return nil, fmt.Errorf("GeocoderAPIAmap.Parse(): [%s] %s", r.Status, r.Info)
	}
	//	Transform
	r0 := r.Geocodes[0]
	///	解析经纬度
	loc := strings.Split(r0.Location, ",")
	var longitude, latitude float64
	if longitude, err = strconv.ParseFloat(loc[0], 64); err != nil {
		return nil, fmt.Errorf("无法解析经纬度：%v: %s", r0, err)
	}
	if latitude, err = strconv.ParseFloat(loc[1], 64); err != nil {
		return nil, fmt.Errorf("无法解析经纬度：%v: %s", r0, err)
	}
	//	坐标转换：GCJ02 => WGS84
	l2 := gocoord.GCJ02ToWGS84(gocoord.Position{Lon: longitude, Lat: latitude})
	// fmt.Printf("GCJ02 (%.6f, %.6f) => WGS84 (%.6f, %6f)\n",
	// 	longitude, latitude,
	// 	l2.Lat, l2.Lon,
	// )

	//	返回
	return &Address{Address: r0.Formatted_Address, Longitude: l2.Lon, Latitude: l2.Lat}, nil
}

// https://lbs.amap.com/api/webservice/guide/api/batchrequest

func (g GeocoderAPIAmap) RequestBatch(addrs []string) ([]Address, error) {
	var err error

	u, err := url.Parse("https://restapi.amap.com/v3/batch?key=" + g.key)
	if err != nil {
		return nil, err
	}

	ops := []GeocoderAPIAmapBatchRequestItem{}
	for _, addr := range addrs {
		u, err := url.Parse("https://restapi.amap.com/v3/geocode/geo")
		if err != nil {
			return nil, err
		}
		params := u.Query()
		params.Add("key", g.key)
		params.Add("address", addr)
		u.RawQuery = params.Encode()

		ops = append(ops, GeocoderAPIAmapBatchRequestItem{u.RequestURI()})
	}
	req := GeocoderAPIAmapBatchRequest{ops}
	post_body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	// fmt.Printf("> [%d] %s\n", len(addrs), post_body)

	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(post_body))
	if err != nil {
		return nil, err
	}

	//	Load Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resps := GeocoderAPIAmapBatchResponse{}

	if err = json.Unmarshal(body, &resps); err != nil {
		return nil, fmt.Errorf("JSON解析失败：'%s' => %s", body, err)
	}

	result := make([]Address, 0, len(addrs))
	if len(resps) == len(addrs) {
		for i, r := range resps {
			if len(r.Body.Geocodes) > 0 {
				r0 := r.Body.Geocodes[0]
				///	解析经纬度
				loc := strings.Split(r0.Location, ",")
				var longitude, latitude float64
				if longitude, err = strconv.ParseFloat(loc[0], 64); err != nil {
					log.Errorf("无法解析经纬度：%v: %s", r0, err)
				}
				if latitude, err = strconv.ParseFloat(loc[1], 64); err != nil {
					log.Errorf("无法解析经纬度：%v: %s", r0, err)
				}
				//	坐标转换：GCJ02 => WGS84
				l2 := gocoord.GCJ02ToWGS84(gocoord.Position{Lon: longitude, Lat: latitude})
				result = append(result, Address{
					Address:   r0.Formatted_Address,
					Longitude: l2.Lon,
					Latitude:  l2.Lat,
				})
			} else {
				//	添加坐标为0的地址
				result = append(result, Address{Address: addrs[i]})
			}
		}
	}

	return result, nil
}
