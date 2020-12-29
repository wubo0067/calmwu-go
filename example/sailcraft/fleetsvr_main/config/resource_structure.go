package config

import (
	"sailcraft/base"
)

const (
	COUNT_TYPE_CONST = "constant"
	COUNT_TYPE_RANGE = "range"
)

const (
	RESOURCE_ITEM_TYPE_GEM       = "gem"      // 钻石
	RESOURCE_ITEM_TYPE_IRON      = "iron"     // 铁矿
	RESOURCE_ITEM_TYPE_STONE     = "stone"    // 石材
	RESOURCE_ITEM_TYPE_WOOD      = "wood"     // 木材
	RESOURCE_ITEM_TYPE_GOLD      = "gold"     // 金币
	RESOURCE_ITEM_TYPE_EXP       = "exp"      // 经验
	RESOURCE_ITEM_TYPE_VITALITY  = "vitality" // 活跃度
	RESOURCE_ITEM_TYPE_SHIP_SOUL = "shipsoul" // 舰魂
)

// 道具
type PropItem struct {
	ProtypeId int         `mapstructure:"protype_id" json:"protype_id"`
	CountType string      `mapstructure:"count_type" json:"count_type"`
	Count     interface{} `mapstructure:"count"      json:"count"`
}

// 战舰卡片
type BattleShipCardItem struct {
	ProtypeId int         `mapstructure:"protype_id" json:"protype_id"`
	CountType string      `mapstructure:"count_type" json:"count_type"`
	Count     interface{} `mapstructure:"count"      json:"count"`
}

// 资源
type ResourceItem struct {
	Type      string      `mapstructure:"type"       json:"type"`
	CountType string      `mapstructure:"count_type" json:"count_type"`
	Count     interface{} `mapstructure:"count"      json:"count"`
}

type ResourcesAttr struct {
	BattleShipCards []BattleShipCardItem `mapstructure:"battleship_cards" json:"battleship_cards"`
	ResourceItems   []ResourceItem       `mapstructure:"resources"        json:"resources"`
	PropItems       []PropItem           `mapstructure:"props"            json:"props"`
}

func (attr *ResourcesAttr) Add(other *ResourcesAttr) {
	addShipCards := make([]*BattleShipCardItem, 0, len(other.BattleShipCards))
	addResources := make([]*ResourceItem, 0, len(other.ResourceItems))
	addProps := make([]*PropItem, 0, len(other.PropItems))

	for i, _ := range other.BattleShipCards {
		addShipCards = append(addShipCards, &other.BattleShipCards[i])
	}

	for i, _ := range other.ResourceItems {
		addResources = append(addResources, &other.ResourceItems[i])
	}

	for i, _ := range other.PropItems {
		addProps = append(addProps, &other.PropItems[i])
	}

	attr.AddShipCards(1, addShipCards...)
	attr.AddResources(1, addResources...)
	attr.AddProps(1, addProps...)
}

func (attr *ResourcesAttr) Scale(factor float64) {

	for index, _ := range attr.BattleShipCards {
		shipCard := &(attr.BattleShipCards[index])
		shipCard.Count = scaleCount(shipCard.Count, factor)
	}

	for index, _ := range attr.ResourceItems {
		resourceItem := &(attr.ResourceItems[index])
		resourceItem.Count = scaleCount(resourceItem.Count, factor)
	}

	for index, _ := range attr.PropItems {
		prop := &(attr.PropItems[index])
		prop.Count = scaleCount(prop.Count, factor)
	}
}

func (attr *ResourcesAttr) Sub(other *ResourcesAttr) {
	addShipCards := make([]*BattleShipCardItem, 0, len(other.BattleShipCards))
	addResources := make([]*ResourceItem, 0, len(other.ResourceItems))
	addProps := make([]*PropItem, 0, len(other.PropItems))

	for i, _ := range other.BattleShipCards {
		addShipCards = append(addShipCards, &other.BattleShipCards[i])
	}

	for i, _ := range other.ResourceItems {
		addResources = append(addResources, &other.ResourceItems[i])
	}

	for i, _ := range other.PropItems {
		addProps = append(addProps, &other.PropItems[i])
	}

	attr.AddShipCards(-1, addShipCards...)
	attr.AddResources(-1, addResources...)
	attr.AddProps(-1, addProps...)
}

func (attr *ResourcesAttr) AddShipCards(scaler float64, shipCardsSlice ...*BattleShipCardItem) {
	shipCards := make(map[int]*BattleShipCardItem)

	for index, _ := range attr.BattleShipCards {
		shipCards[attr.BattleShipCards[index].ProtypeId] = &(attr.BattleShipCards[index])
	}

	for _, shipCard := range shipCardsSlice {
		if currentShipCard, ok := shipCards[shipCard.ProtypeId]; ok {
			minCount, maxCount := addCount(currentShipCard.Count, scaleCount(shipCard.Count, scaler))
			if minCount == maxCount {
				currentShipCard.CountType = COUNT_TYPE_CONST
				currentShipCard.Count = minCount
			} else {
				currentShipCard.CountType = COUNT_TYPE_RANGE
				currentShipCard.Count = []float64{minCount, maxCount}
			}
		} else {
			newShipCard := BattleShipCardItem{ProtypeId: shipCard.ProtypeId, CountType: shipCard.CountType, Count: scaleCount(shipCard.Count, scaler)}
			attr.BattleShipCards = append(attr.BattleShipCards, newShipCard)
			shipCards[newShipCard.ProtypeId] = &newShipCard
		}
	}
}

func (attr *ResourcesAttr) AddResources(scaler float64, resourcesSlice ...*ResourceItem) {
	resources := make(map[string]*ResourceItem)

	for index, _ := range attr.ResourceItems {
		resources[attr.ResourceItems[index].Type] = &(attr.ResourceItems[index])
	}

	for _, resourceItem := range resourcesSlice {
		if currentItem, ok := resources[resourceItem.Type]; ok {
			minCount, maxCount := addCount(currentItem.Count, scaleCount(resourceItem.Count, scaler))

			if minCount == maxCount {
				currentItem.CountType = COUNT_TYPE_CONST
				currentItem.Count = minCount
			} else {
				currentItem.CountType = COUNT_TYPE_RANGE
				currentItem.Count = []float64{minCount, maxCount}
			}

		} else {
			newResourceItem := ResourceItem{Type: resourceItem.Type, CountType: resourceItem.CountType, Count: scaleCount(resourceItem.Count, scaler)}
			attr.ResourceItems = append(attr.ResourceItems, newResourceItem)
			resources[newResourceItem.Type] = &newResourceItem
		}
	}
}

func (attr *ResourcesAttr) AddProps(scaler float64, propSlice ...*PropItem) {
	props := make(map[int]*PropItem)

	for index, _ := range attr.PropItems {
		props[attr.PropItems[index].ProtypeId] = &(attr.PropItems[index])
	}

	for _, prop := range propSlice {
		if currentProp, ok := props[prop.ProtypeId]; ok {
			minCount, maxCount := addCount(currentProp.Count, scaleCount(prop.Count, scaler))
			if minCount == maxCount {
				currentProp.CountType = COUNT_TYPE_CONST
				currentProp.Count = minCount
			} else {
				currentProp.CountType = COUNT_TYPE_RANGE
				currentProp.Count = []float64{minCount, maxCount}
			}
		} else {
			newProp := PropItem{ProtypeId: prop.ProtypeId, CountType: prop.CountType, Count: scaleCount(prop.Count, scaler)}
			attr.PropItems = append(attr.PropItems, newProp)
			props[newProp.ProtypeId] = &newProp
		}
	}
}

func (attr *ResourcesAttr) EscapeNil() {
	if attr.BattleShipCards == nil {
		attr.BattleShipCards = make([]BattleShipCardItem, 0)
	}

	if attr.PropItems == nil {
		attr.PropItems = make([]PropItem, 0)
	}

	if attr.ResourceItems == nil {
		attr.ResourceItems = make([]ResourceItem, 0)
	}
}

func (attr *ResourcesAttr) GetResourceItem(key string) *ResourceItem {
	for _, item := range attr.ResourceItems {
		if item.Type == key {
			return &item
		}
	}

	return nil
}

func scaleCount(c interface{}, factor float64) interface{} {
	if value, ok := c.(float64); ok {
		return value * factor
	} else if sliceValue, ok := c.([]interface{}); ok {
		sliceLen := len(sliceValue)
		ret := make([]interface{}, sliceLen)
		for i, v := range sliceValue {
			ret[i] = v.(float64) * factor
		}

		return ret
	} else {
		return 0
	}
}

func addCount(c1 interface{}, c2 interface{}) (minCount float64, maxCount float64) {
	minCount = 0
	maxCount = 0

	if value, ok := c1.(float64); ok {
		minCount = value
		maxCount = value
	} else if sliceValue, ok := c1.([]interface{}); ok {
		sliceLen := len(sliceValue)
		if sliceLen >= 2 {
			minCount = sliceValue[0].(float64)
			maxCount = sliceValue[1].(float64)
		} else if sliceLen == 1 {
			minCount = sliceValue[0].(float64)
			maxCount = sliceValue[0].(float64)
		}
	}

	if value, ok := c2.(float64); ok {
		minCount += value
		maxCount += value
	} else if sliceValue, ok := c2.([]interface{}); ok {
		sliceLen := len(sliceValue)
		if sliceLen >= 2 {
			minCount += sliceValue[0].(float64)
			maxCount += sliceValue[1].(float64)
		} else if sliceLen == 1 {
			minCount += sliceValue[0].(float64)
			maxCount += sliceValue[0].(float64)
		}
	}

	return
}

func ConstantValue(c interface{}) float64 {
	return base.ConvertToFloat64(c, 0.0)
}

func MinMaxValue(c interface{}) (min float64, max float64) {
	min = 0
	max = 0

	if sValue, ok := c.([]interface{}); ok {
		sLen := len(sValue)
		if sLen >= 2 {
			min = base.ConvertToFloat64(sValue[0], 0.0)
			max = base.ConvertToFloat64(sValue[1], 0.0)
		}
	}

	return
}
