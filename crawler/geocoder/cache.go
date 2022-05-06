package geocoder

import (
	"bytes"
	"encoding/gob"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/syndtr/goleveldb/leveldb"
)

type GeocodeCache struct {
	db *leveldb.DB
}

func NewGeocodeCache(path string) (*GeocodeCache, error) {
	db, err := leveldb.OpenFile(path, nil)
	if err != nil {
		return nil, err
	}

	log.Infof("成功建立缓存目录：%s", path)

	c := GeocodeCache{db: db}
	return &c, nil
}

func (c GeocodeCache) Put(addr string, longitude, latitude float64) error {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode([]float64{longitude, latitude}); err != nil {
		return err
	}
	if err := c.db.Put([]byte(strings.TrimSpace(addr)), buf.Bytes(), nil); err != nil {
		return err
	}
	return nil
}

func (c GeocodeCache) Get(addr string) (longitude, latitude float64, err error) {
	var data []byte
	if data, err = c.db.Get([]byte(strings.TrimSpace(addr)), nil); err != nil {
		return 0, 0, err
	}
	loc := []float64{}
	if err := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&loc); err != nil {
		return 0, 0, err
	}
	longitude, latitude = loc[0], loc[1]
	return
}
