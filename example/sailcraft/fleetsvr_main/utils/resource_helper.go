package utils

import (
	"fmt"
	"math"
	"sailcraft/base"
	"sailcraft/fleetsvr_main/config"
)

func GemToResource(count int, resourceType string) (int, error) {
	if count <= 0 {
		return 0, nil
	}

	gemExchangeAttr, err := config.GResourceGemExchangeConfig.GetGemExchangeAttr()
	if err != nil {
		return 0, err
	}

	var resourceSlice []int
	switch resourceType {
	case config.RESOURCE_ITEM_TYPE_GOLD:
		resourceSlice = gemExchangeAttr.GoldRanges
	case config.RESOURCE_ITEM_TYPE_WOOD:
		resourceSlice = gemExchangeAttr.WoodRanges
	case config.RESOURCE_ITEM_TYPE_STONE:
		resourceSlice = gemExchangeAttr.StoneRanges
	case config.RESOURCE_ITEM_TYPE_IRON:
		resourceSlice = gemExchangeAttr.IronRanges
	default:
		return -1, fmt.Errorf("resource type is error")
	}

	gemSlice := gemExchangeAttr.GemRanges

	if len(resourceSlice) < 2 || len(gemSlice) < 2 {
		return 0, fmt.Errorf("config length not enough")
	}

	if count <= resourceSlice[0] {
		return gemSlice[0], nil
	} else {
		var index int = len(resourceSlice) - 1
		for idx, value := range resourceSlice {
			if value > count {
				index = idx
				break
			}
		}

		base.GLog.Debug("count %d last_resourceSlice %d resourceSlice %d last_gemSlice %d gemSlice %d", count,
			resourceSlice[index-1], resourceSlice[index], gemSlice[index-1], gemSlice[index])

		var overIndexResource float64 = float64(count - resourceSlice[index-1])
		var overIndexGem float64 = 1.0
		var deltaResource float64 = float64(resourceSlice[index] - resourceSlice[index-1])
		var deltaGem float64 = float64(gemSlice[index] - gemSlice[index-1])
		if deltaGem > 0 && deltaResource > 0 {
			overIndexGem = overIndexResource * deltaGem / deltaResource
		}

		base.GLog.Debug("overIndexResource %f deltaResource %f deltaGem %f overIndexGem %f", overIndexResource,
			deltaResource, deltaGem, overIndexGem)

		realGemCount := int(math.Floor(overIndexGem)) + gemSlice[index-1]

		return realGemCount, nil
	}
}
