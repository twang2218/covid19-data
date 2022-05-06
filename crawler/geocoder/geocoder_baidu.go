package geocoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"github.com/suifengtec/gocoord"
)

// https://lbsyun.baidu.com/index.php?title=webapi/guide/changeposition
// https://lbsyun.baidu.com/index.php?title=coordinate

// 百度地图
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
type GeocoderAPIBaidu struct {
	key string
}

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

type GeocoderAPIBaiduBatchRequest struct {
	Reqs []GeocoderAPIBaiduBatchRequestItem `json:"reqs"`
}

type GeocoderAPIBaiduBatchRequestItem struct {
	Method string `json:"method"`
	Url    string `json:"url"`
}

type GeocoderAPIBaiduBatchResponse struct {
	Status      int                        `json:"status"`
	BatchResult []GeocoderAPIBaiduResponse `json:"batch_result"`
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

func (a GeocoderAPIBaidu) Name() string {
	return "百度地图API"
}

func (a GeocoderAPIBaidu) Request(addr string) (*Address, error) {
	var err error

	u, err := url.Parse("https://api.map.baidu.com/geocoding/v3/")
	if err != nil {
		return nil, err
	}

	params := u.Query()
	params.Add("address", addr)
	params.Add("ret_coordtype", "gcj02ll") // gcj02ll: 国测局坐标; bd09mc: 百度墨卡托坐标; bd09ll: 百度经纬度坐标
	params.Add("output", "json")
	params.Add("ak", a.key)
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

	//	解析 JSON
	r := GeocoderAPIBaiduResponse{}

	if err = json.Unmarshal(body, &r); err != nil {
		return nil, err
	}
	//	处理API返回结果
	if r.Status != 0 {
		return nil, fmt.Errorf("GeocoderAPIBaidu.Parse(): [%d]", r.Status)
	}
	//	坐标转换：GCJ02 => WGS84
	l2 := gocoord.GCJ02ToWGS84(gocoord.Position{Lon: r.Result.Location.Lng, Lat: r.Result.Location.Lat})
	// fmt.Printf("GCJ02 (%.6f, %.6f) => WGS84 (%.6f, %6f)\n",
	// 	r.Result.Location.Lat, r.Result.Location.Lng,
	// 	l2.Lat, l2.Lon,
	// )
	//	返回
	return &Address{Address: addr, Longitude: l2.Lon, Latitude: l2.Lat}, nil
}

// https://lbsyun.baidu.com/index.php?title=webapi/guide/batch

func (a GeocoderAPIBaidu) RequestBatch(addrs []string) ([]Address, error) {
	var err error

	u, err := url.Parse("https://api.map.baidu.com/batch")
	if err != nil {
		return nil, err
	}

	reqs := []GeocoderAPIBaiduBatchRequestItem{}
	for _, addr := range addrs {
		u, err := url.Parse("https://api.map.baidu.com/geocoding/v3/")
		if err != nil {
			return nil, err
		}
		params := u.Query()
		params.Add("address", addr)
		params.Add("ret_coordtype", "gcj02ll") // gcj02ll: 国测局坐标; bd09mc: 百度墨卡托坐标; bd09ll: 百度经纬度坐标
		params.Add("output", "json")
		params.Add("ak", a.key)
		u.RawQuery = params.Encode()

		reqs = append(reqs, GeocoderAPIBaiduBatchRequestItem{Method: "get", Url: u.RequestURI()})
	}
	req := GeocoderAPIBaiduBatchRequest{reqs}

	post_body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(u.String(), "application/json", bytes.NewBuffer(post_body))
	if err != nil {
		return nil, err
	}

	//	Load Response Body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	resps := GeocoderAPIBaiduBatchResponse{}

	if err = json.Unmarshal(body, &resps); err != nil {
		return nil, fmt.Errorf("JSON解析失败：'%s' => %s", body, err)
	}

	result := make([]Address, 0, len(addrs))
	if resps.Status == 0 && len(resps.BatchResult) == len(addrs) {
		for i, r := range resps.BatchResult {
			if r.Status == 0 {
				//	坐标转换：GCJ02 => WGS84
				l2 := gocoord.GCJ02ToWGS84(gocoord.Position{Lon: r.Result.Location.Lng, Lat: r.Result.Location.Lat})
				// fmt.Printf("GCJ02 (%.6f, %.6f) => WGS84 (%.6f, %6f)\n",
				// 	resp.Result.Location.Lat, resp.Result.Location.Lng,
				// 	l2.Lat, l2.Lon,
				// )
				//	添加解析结果
				result = append(result, Address{
					Address:   addrs[i],
					Longitude: l2.Lon,
					Latitude:  l2.Lat,
				})
			} else {
				//	无法解析，添加坐标为0的地址
				result = append(result, Address{Address: addrs[i]})
			}
		}
	}

	return result, nil
}
