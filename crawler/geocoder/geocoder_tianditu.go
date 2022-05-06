package geocoder

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	log "github.com/sirupsen/logrus"
)

// http://lbs.tianditu.gov.cn/server/geocodinginterface.html
// https://console.tianditu.gov.cn/api/statistics

// 天地图
//	[Request]
//		https://api.tianditu.gov.cn/geocoder?ds={"keyWord":"上海市徐汇区斜土路2365弄"}&tk=您的密钥
//	[Response]
// {
// 	"msg": "ok",
// 	"location": {
// 	  "score": 95,
// 	  "level": "地名地址",
// 	  "lon": 121.44344,
// 	  "lat": 31.194107,
// 	  "keyWord": "上海市徐汇区斜土路2365弄"
// 	},
// 	"searchVersion": "6.0.0",
// 	"status": "0"
// }

type GeocoderAPITiandituResponse struct {
	MSG           string
	SearchVersion string
	Status        string
	Location      struct {
		Score   int
		Level   string
		Lon     float64
		Lat     float64
		Keyword string
	}
}

type GeocoderAPITianditu struct {
	key string
}

func NewGeocoderTianditu(key, cachedir string) Geocoder {
	var cache *GeocodeCache
	if len(cachedir) > 0 {
		var err error
		cache, err = NewGeocodeCache(cachedir)
		if err != nil {
			log.Errorf("NewGeocoderTianditu(): 无法建立缓存[%s]：%s", cachedir, err)
		}
	}
	return Geocoder{api: GeocoderAPITianditu{key: key}, cache: cache}
}

func (a GeocoderAPITianditu) Name() string {
	return "天地图API"
}

func (a GeocoderAPITianditu) Request(addr string) (*Address, error) {
	var err error

	u, err := url.Parse("https://api.tianditu.gov.cn/geocoder")
	if err != nil {
		return nil, err
	}
	params := u.Query()
	params.Add("tk", a.key)
	params.Add("ds", strings.ReplaceAll("{\"keyWord\":\"{addr}\"}", "{addr}", addr))
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
	r := GeocoderAPITiandituResponse{}

	if err = json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("JSON解析失败：'%s' => %s", body, err)
	}
	//	处理API返回结果
	if r.Status != "0" {
		if len(r.MSG) > 0 {
			return nil, fmt.Errorf("GeocoderAPITianditu.Parse(): [%s] %s", r.Status, r.MSG)
		} else {
			return nil, fmt.Errorf("GeocoderAPITianditu.Parse(): %s", body)
		}
	}
	//	天地图的坐标系接近 WGS84，所以不进行转换

	//	返回
	return &Address{Address: r.Location.Keyword, Longitude: r.Location.Lon, Latitude: r.Location.Lat}, nil
}

// 天地图没有批处理API，因此对单次请求进行封装
func (a GeocoderAPITianditu) RequestBatch(addrs []string) ([]Address, error) {
	result := make([]Address, 0, len(addrs))
	for _, addr := range addrs {
		if a, err := a.Request(addr); err == nil {
			//	添加解析结果
			result = append(result, *a)
		} else {
			//	无法解析，添加坐标为0的地址
			result = append(result, Address{Address: addr})
		}
	}
	return result, nil
}
