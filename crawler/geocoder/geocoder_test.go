package geocoder

import (
	"fmt"
	"math"
	"os"
	"path"
	"testing"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Errorf("Cannot load .env file. %q", err)
	}
}

const THRESHOLD = 0.01

func TestGeocoder(t *testing.T) {
	testcases := []Address{
		{"上海市静安区芷江西路453弄", 121.45280, 31.25884}, // 31.25884,121.45280
		{"上海市浦东新区微山路", 121.50859, 31.21077},
		{"山东省青岛市胶州市皓月路", 120.02190, 36.27371},
	}
	addrs := []string{}
	for _, c := range testcases {
		addrs = append(addrs, c.Address)
	}

	gs := []Geocoder{
		NewGeocoderAMAP(os.Getenv("KEY_AMAP"), path.Join(os.TempDir(), "geocoder", "amap")),
		NewGeocoderTianditu(os.Getenv("KEY_TIANDITU"), path.Join(os.TempDir(), "geocoder", "tianditu")),
		NewGeocoderBaidu(os.Getenv("KEY_BAIDU_MAP"), path.Join(os.TempDir(), "geocoder", "baidu")),
	}

	for _, g := range gs {
		//	单次请求
		for i, c := range testcases {
			a, err := g.Geocode(c.Address)
			assert.NoError(t, err, fmt.Sprintf("%s: 返回错误：(%d) %q => %s", g.Name(), i, c.Address, err))
			if err == nil {
				assert.Less(t, math.Abs(a.Longitude-c.Longitude), THRESHOLD,
					fmt.Sprintf("%s: 解析经度误差过大 %q (%.5f) => (%.5f)", g.Name(), c.Address, c.Longitude, a.Longitude))
				assert.Less(t, math.Abs(a.Latitude-c.Latitude), THRESHOLD,
					fmt.Sprintf("%s: 解析纬度误差过大 %q (%.5f) => (%.5f)", g.Name(), c.Address, c.Latitude, a.Latitude))
			}
		}

		//	批量请求
		as, err := g.GeocodeInBatch(addrs)
		assert.NoError(t, err, fmt.Sprintf("%s: 批处理返回错误：%s", g.Name(), err))
		if err == nil {
			for i, c := range testcases {
				assert.Less(t, math.Abs(as[i].Longitude-c.Longitude), THRESHOLD,
					fmt.Sprintf("%s: 批量处理 - 解析经度误差过大 %q (%.5f) => (%.5f)", g.Name(), c.Address, c.Longitude, as[i].Longitude))
				assert.Less(t, math.Abs(as[i].Latitude-c.Latitude), THRESHOLD,
					fmt.Sprintf("%s: 批量处理 - 解析纬度误差过大 %q (%.5f) => (%.5f)", g.Name(), c.Address, c.Latitude, as[i].Latitude))
			}
		}
	}

}
