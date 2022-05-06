package geocoder

import (
	"math"

	log "github.com/sirupsen/logrus"
)

type Address struct {
	Address   string
	Longitude float64
	Latitude  float64
}

type Geocoder struct {
	api   GeocoderAPI
	cache *GeocodeCache
}

func (g Geocoder) Name() string {
	return g.api.Name()
}

func (g Geocoder) Geocode(addr string) (*Address, error) {
	//	先检查缓存是否已存在该地址的解析
	if g.cache != nil {
		lon, lat, err := g.cache.Get(addr)
		if err != nil {
			// log.Warnf("Geocode(%s): 读取失败：%s", addr, err)
		} else {
			return &Address{Address: addr, Longitude: lon, Latitude: lat}, nil
		}
	}
	//	缓存没有，发起请求
	a, err := g.api.Request(addr)
	if err != nil {
		return a, err
	}
	//	将结果非 0 的坐标信息保存于缓存
	if g.cache != nil && a.Longitude != 0 && a.Latitude != 0 {
		err := g.cache.Put(a.Address, a.Longitude, a.Latitude)
		if err != nil {
			log.Errorf("Geocode(%q): 写入失败：%s", addr, err)
		}
	}
	return a, nil
}

func (g Geocoder) GeocodeInBatch(addrs []string) ([]Address, error) {
	results := make([]Address, len(addrs))
	//	先检查缓存是否已存在该地址的解析,分拆已经解析和未被解析的地址列表
	addrs_to_query := make([]string, 0, len(addrs))
	id_to_query := make([]int, 0, len(addrs))
	for i, addr := range addrs {
		if g.cache != nil {
			if lon, lat, err := g.cache.Get(addr); err == nil {
				//  缓存查询成功，将结果存入结果
				results[i].Address = addr
				results[i].Longitude = lon
				results[i].Latitude = lat
				//	跳过后面添加查询列表
				continue
			}
		}
		//	缓存查询失败(或无缓存)，将地址添加到查询列表，并保存对应ID
		addrs_to_query = append(addrs_to_query, addr)
		id_to_query = append(id_to_query, i)
	}

	// fmt.Printf("addrs: (%d) => addrs_to_query: (%d)\n", len(addrs), len(addrs_to_query))

	//	针对未被解析的地址列表进行批量解析
	var results_from_query []Address
	if len(addrs_to_query) > 0 {
		//	根据限制切片，分子批发送请求
		addrs_to_query_slices := batch_split(addrs)
		for _, addrs_slice := range addrs_to_query_slices {
			results_slice, err := g.api.RequestBatch(addrs_slice)
			if err == nil {
				//	请求成功，追加结果
				results_from_query = append(results_from_query, results_slice...)
			} else {
				//	请求失败，追加空坐标到地址列表
				results_from_query = append(results_from_query, (make([]Address, len(addrs_slice)))...)
			}
		}

		//	将结果非 0 的坐标信息保存于缓存
		if g.cache != nil {
			for _, a := range results_from_query {
				if a.Longitude != 0 && a.Latitude != 0 {
					if err := g.cache.Put(a.Address, a.Longitude, a.Latitude); err != nil {
						log.Errorf("Geocode(%q): 缓存写入失败：%s", a.Address, err)
					}
				}
			}
		}
	}

	//	合并之前拆分的已解析列表与未被解析的地址列表，需保持原有顺序
	for i, result := range results_from_query {
		id := id_to_query[i]
		results[id] = result
	}
	//	返回结果
	return results, nil
}

func (g Geocoder) Close() {
	if g.cache != nil {
		g.cache.db.Close()
	}
}

type GeocoderAPI interface {
	Name() string
	Request(addr string) (*Address, error)
	RequestBatch(addrs []string) ([]Address, error)
}

const BATCH_SIZE_LIMIT int = 20

func batch_split(addrs []string) [][]string {
	num_of_slice := int(math.Ceil(float64(len(addrs)) / float64(BATCH_SIZE_LIMIT)))
	addrs_slices := make([][]string, 0)
	var i int
	for i = 0; i < num_of_slice; i++ {
		addrs_slices = append(addrs_slices, addrs[i*BATCH_SIZE_LIMIT:(i+1)*BATCH_SIZE_LIMIT])
	}
	return addrs_slices
}
