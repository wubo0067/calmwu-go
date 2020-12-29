package handler

import (
	"math/rand"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
	"sailcraft/fleetsvr_main/proto"
)

func ResourcesConfigToProto(src *config.ResourcesAttr, dst *proto.ProtoResourcesAttr) {
	if src == nil && dst == nil {
		return
	}

	r := rand.New(rand.NewSource(base.GLocalizedTime.SecTimeStamp()))

	shipCardLen := len(src.BattleShipCards)
	dst.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, 0, shipCardLen)
	for _, shipCard := range src.BattleShipCards {
		count := 0
		if shipCard.CountType == config.COUNT_TYPE_RANGE {
			min, max := config.MinMaxValue(shipCard.Count)
			count = int(min) + r.Intn(int(max-min))
		} else {
			count = int(config.ConstantValue(shipCard.Count))
		}

		protoCard := new(proto.ProtoBattleShipCardItem)
		protoCard.ProtypeId = shipCard.ProtypeId
		protoCard.CountType = config.COUNT_TYPE_CONST
		protoCard.Count = count
		dst.BattleShipCards = append(dst.BattleShipCards, protoCard)
	}

	resLen := len(src.ResourceItems)
	dst.ResourceItems = make([]*proto.ProtoResourceItem, 0, resLen)
	for _, resItem := range src.ResourceItems {
		count := 0
		if resItem.CountType == config.COUNT_TYPE_RANGE {
			min, max := config.MinMaxValue(resItem.Count)
			count = int(min) + r.Intn(int(max-min))
		} else {
			count = int(config.ConstantValue(resItem.Count))
		}

		protoRes := new(proto.ProtoResourceItem)
		protoRes.Type = resItem.Type
		protoRes.CountType = config.COUNT_TYPE_CONST
		protoRes.Count = count
		dst.ResourceItems = append(dst.ResourceItems, protoRes)
	}

	propLen := len(src.PropItems)
	dst.PropItems = make([]*proto.ProtoPropItem, 0, propLen)
	for _, propItem := range src.PropItems {
		count := 0
		if propItem.CountType == config.COUNT_TYPE_RANGE {
			min, max := config.MinMaxValue(propItem.Count)
			count = int(min) + r.Intn(int(max-min))
		} else {
			count = int(config.ConstantValue(propItem.Count))
		}

		protoProp := new(proto.ProtoPropItem)
		protoProp.ProtypeId = propItem.ProtypeId
		protoProp.CountType = config.COUNT_TYPE_CONST
		protoProp.Count = count
		dst.PropItems = append(dst.PropItems, protoProp)
	}
}

func ResourcesConfigToProtoWithOmit(src *config.ResourcesAttr, dst *proto.ProtoResourcesAttr, omitResType ...string) {
	if src == nil && dst == nil {
		return
	}

	shipCardLen := len(src.BattleShipCards)
	dst.BattleShipCards = make([]*proto.ProtoBattleShipCardItem, shipCardLen)
	for i, shipCard := range src.BattleShipCards {
		protoCard := new(proto.ProtoBattleShipCardItem)
		protoCard.ProtypeId = shipCard.ProtypeId
		protoCard.CountType = config.COUNT_TYPE_CONST
		protoCard.Count = int(config.ConstantValue(shipCard.Count))
		dst.BattleShipCards[i] = protoCard
	}

	dst.ResourceItems = make([]*proto.ProtoResourceItem, 0)
	for _, resItem := range src.ResourceItems {
		omit := false
		for _, resType := range omitResType {
			if resItem.Type == resType {
				omit = true
				break
			}
		}

		if omit {
			continue
		}

		protoRes := new(proto.ProtoResourceItem)
		protoRes.Type = resItem.Type
		protoRes.CountType = config.COUNT_TYPE_CONST
		protoRes.Count = int(config.ConstantValue(resItem.Count))
		dst.ResourceItems = append(dst.ResourceItems, protoRes)
	}

	propLen := len(src.PropItems)
	dst.PropItems = make([]*proto.ProtoPropItem, propLen)
	for i, propItem := range src.PropItems {
		protoProp := new(proto.ProtoPropItem)
		protoProp.ProtypeId = propItem.ProtypeId
		protoProp.CountType = config.COUNT_TYPE_CONST
		protoProp.Count = int(config.ConstantValue(propItem.Count))
		dst.PropItems[i] = protoProp
	}
}

func AppendProtoResources(dst *proto.ProtoResourcesAttr, other *proto.ProtoResourcesAttr) *proto.ProtoResourcesAttr {
	shipCards := make(map[int]*proto.ProtoBattleShipCardItem)
	resources := make(map[string]*proto.ProtoResourceItem)
	props := make(map[int]*proto.ProtoPropItem)

	for index, _ := range dst.BattleShipCards {
		shipCards[dst.BattleShipCards[index].ProtypeId] = dst.BattleShipCards[index]
	}

	for index, _ := range dst.ResourceItems {
		resources[dst.ResourceItems[index].Type] = dst.ResourceItems[index]
	}

	for index, _ := range dst.PropItems {
		props[dst.PropItems[index].ProtypeId] = dst.PropItems[index]
	}

	// 合并船卡
	for _, shipCard := range other.BattleShipCards {
		if currentShipCard, ok := shipCards[shipCard.ProtypeId]; ok {
			currentShipCard.Count += shipCard.Count
		} else {
			newShipCard := &proto.ProtoBattleShipCardItem{ProtypeId: shipCard.ProtypeId, CountType: config.COUNT_TYPE_CONST, Count: shipCard.Count}
			dst.BattleShipCards = append(dst.BattleShipCards, newShipCard)
			shipCards[newShipCard.ProtypeId] = newShipCard
		}
	}

	// 合并资源
	for _, resourceItem := range other.ResourceItems {
		if currentItem, ok := resources[resourceItem.Type]; ok {
			currentItem.Count += resourceItem.Count
		} else {
			newResourceItem := &proto.ProtoResourceItem{Type: resourceItem.Type, CountType: config.COUNT_TYPE_CONST, Count: resourceItem.Count}
			dst.ResourceItems = append(dst.ResourceItems, newResourceItem)
			resources[newResourceItem.Type] = newResourceItem
		}
	}

	// 合并道具
	for _, prop := range other.PropItems {
		if currentProp, ok := props[prop.ProtypeId]; ok {
			currentProp.Count += prop.Count
		} else {
			newProp := &proto.ProtoPropItem{ProtypeId: prop.ProtypeId, CountType: config.COUNT_TYPE_CONST, Count: prop.Count}
			dst.PropItems = append(dst.PropItems, newProp)
			props[newProp.ProtypeId] = newProp
		}
	}

	return dst
}
