package geocoder

import (
	"math"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

const TEST_CACHE_THRESHOLD = 0.001

func TestCache(t *testing.T) {
	testcases := []Address{
		{"上海市静安区芷江西路453弄", 121.45779, 31.25999},
	}

	cache, err := NewGeocodeCache(path.Join(os.TempDir(), "geocoder", "cache"))
	assert.NoErrorf(t, err, "建立缓存失败: %s", err)
	if err == nil {
		for _, c := range testcases {
			err := cache.Put(c.Address, c.Longitude, c.Latitude)
			assert.NoError(t, err, "缓存添加地址失败")
		}

		for _, c := range testcases {
			longitude, latitude, err := cache.Get(c.Address)
			assert.NoErrorf(t, err, "获取缓存地址失败")
			assert.Lessf(t, math.Abs(longitude-c.Longitude), TEST_CACHE_THRESHOLD, "缓存返回经度超出误差：%f => %f", c.Longitude, longitude)
			assert.Lessf(t, math.Abs(latitude-c.Latitude), TEST_CACHE_THRESHOLD, "缓存返回纬度超出误差：%f => %f", c.Latitude, latitude)
		}
	}
}
