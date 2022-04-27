package geocoder

import (
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
		{"上海市静安区芷江西路453弄", 121.45779, 31.25999},
	}

	var g Geocoder

	g = NewGeocoderAMAP(os.Getenv("KEY_AMAP"), path.Join(os.TempDir(), "geocoder", "amap"))
	for i, c := range testcases {
		gi, err := g.Geocode(c.Address)
		assert.NoErrorf(t, err, "高德地图API： 返回错误：(%d) %q => %s", i, c, err)
		if err == nil {
			assert.Less(t, math.Abs(gi.Longitude-c.Longitude), THRESHOLD)
			assert.Less(t, math.Abs(gi.Latitude-c.Latitude), THRESHOLD)
		}
	}

	g = NewGeocoderTianditu(os.Getenv("KEY_TIANDITU"), path.Join(os.TempDir(), "geocoder", "tianditu"))
	for i, c := range testcases {
		gi, err := g.Geocode(c.Address)
		assert.NoErrorf(t, err, "天地图API： 返回错误：(%d) %q => %s", i, c, err)
		if err == nil {
			assert.Less(t, math.Abs(gi.Longitude-c.Longitude), THRESHOLD)
			assert.Less(t, math.Abs(gi.Latitude-c.Latitude), THRESHOLD)
		}
	}

	g = NewGeocoderBaidu(os.Getenv("KEY_BAIDU_MAP"), path.Join(os.TempDir(), "geocoder", "baidu"))
	for i, c := range testcases {
		gi, err := g.Geocode(c.Address)
		assert.NoErrorf(t, err, "百度地图API： 返回错误：(%d) %q => %s", i, c, err)
		if err == nil {
			assert.Less(t, math.Abs(gi.Longitude-c.Longitude), THRESHOLD)
			assert.Less(t, math.Abs(gi.Latitude-c.Latitude), THRESHOLD)
		}
	}
}
