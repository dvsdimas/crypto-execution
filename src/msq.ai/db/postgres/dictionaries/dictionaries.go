package dictionaries

import (
	"github.com/vishalkuo/bimap"
)

type Dictionary interface {
	GetNameById(id int16) string
	GetIdByName(name string) int16
}

type dictionary struct {
	bm *bimap.BiMap
}

func (d dictionary) GetNameById(id int16) string {

	val, _ := d.bm.Get(id)

	return val.(string)
}

func (d dictionary) GetIdByName(name string) int16 {

	val, _ := d.bm.GetInverse(name)

	return val.(int16)
}

type Dictionaries struct {
	exchanges       *Dictionary
	directions      *Dictionary
	orderTypes      *Dictionary
	timeInForce     *Dictionary
	executionTypes  *Dictionary
	executionStatus *Dictionary
}

func (d *Dictionaries) Exchanges() Dictionary {
	return *d.exchanges
}

func (d *Dictionaries) Directions() Dictionary {
	return *d.directions
}

func (d *Dictionaries) OrderTypes() Dictionary {
	return *d.orderTypes
}

func (d *Dictionaries) TimeInForces() Dictionary {
	return *d.timeInForce
}

func (d *Dictionaries) ExecutionTypes() Dictionary {
	return *d.executionTypes
}

func (d *Dictionaries) ExecutionStatuses() Dictionary {
	return *d.executionStatus
}

func NewDictionaries(exchanges *bimap.BiMap,
	directions *bimap.BiMap,
	orderTypes *bimap.BiMap,
	timeInForce *bimap.BiMap,
	executionTypes *bimap.BiMap,
	executionStatus *bimap.BiMap) *Dictionaries {

	var (
		exc, dirs, ots, tif, ets, es Dictionary
	)

	exc = dictionary{bm: exchanges}
	dirs = dictionary{bm: directions}
	ots = dictionary{bm: orderTypes}
	tif = dictionary{bm: timeInForce}
	ets = dictionary{bm: executionTypes}
	es = dictionary{bm: executionStatus}

	return &Dictionaries{
		exchanges:       &exc,
		directions:      &dirs,
		orderTypes:      &ots,
		timeInForce:     &tif,
		executionTypes:  &ets,
		executionStatus: &es,
	}

}
