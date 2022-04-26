package geocoder

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
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

func NewGeocoderTianditu(key string) Geocoder {
	return Geocoder{api: GeocoderAPITianditu{key: key}}
}

func (a GeocoderAPITianditu) GetURL(addr string) *url.URL {
	u, err := url.Parse("https://api.tianditu.gov.cn/geocoder")
	if err != nil {
		return nil
	}
	params := u.Query()
	params.Add("tk", a.key)
	params.Add("ds", strings.ReplaceAll("{\"keyWord\":\"{addr}\"}", "{addr}", addr))
	u.RawQuery = params.Encode()

	return u
}

func (a GeocoderAPITianditu) Parse(body []byte) (*Address, error) {
	gi := Address{}
	//	Parse JSON
	r := GeocoderAPITiandituResponse{}

	var err error
	if err = json.Unmarshal(body, &r); err != nil {
		return &gi, err
	}
	//	Transform
	resp := GeocoderAPITiandituResponse{}
	json.Unmarshal(body, &resp)
	if resp.Status != "0" {
		return nil, fmt.Errorf("GeocoderAPITianditu.Parse(): [%s] %s", resp.Status, resp.MSG)
	}
	//	返回
	return &Address{
		Address:   resp.Location.Keyword,
		Longitude: resp.Location.Lon,
		Latitude:  resp.Location.Lat,
	}, nil
}
