package statistics

import (
	"fmt"
	"strconv"
	"team.gg-server/service"
)

func GetLowerDepthItems(itemId int) ([]int, error) {
	item, exists := service.Items[itemId]
	if !exists {
		return nil, fmt.Errorf("item not found: %d", itemId)
	}
	lowerItems := make([]int, 0)
	if item.Depth == nil || *item.Depth <= 1 {
		return lowerItems, nil
	}
	if item.From == nil || len(*item.From) == 0 {
		return lowerItems, nil
	}
	for _, fromIdStr := range *item.From {
		fromId, err := strconv.Atoi(fromIdStr)
		if err != nil {
			return nil, err
		}
		lowerItems = append(lowerItems, fromId)
	}
	return lowerItems, nil
}

func GetDepth1Items(itemId int) ([]int, error) {
	item, exists := service.Items[itemId]
	if !exists {
		return nil, fmt.Errorf("item not found: %d", itemId)
	}
	selfItems := []int{itemId}
	if item.Depth == nil || *item.Depth <= 1 {
		return selfItems, nil
	}
	if item.From == nil || len(*item.From) == 0 {
		return selfItems, nil
	}
	subItems, err := GetLowerDepthItems(itemId)
	if err != nil {
		return nil, err
	}
	lowerDepthItems := make([]int, 0)
	for _, subItem := range subItems {
		depth1Items, err := GetDepth1Items(subItem)
		if err != nil {
			return nil, err
		}
		lowerDepthItems = append(lowerDepthItems, depth1Items...)
	}
	lowerItemSet := make(map[int]bool)
	for _, lowerItem := range lowerDepthItems {
		lowerItemSet[lowerItem] = true
	}
	uniqueLowerDepthItems := make([]int, 0)
	for lowerItem := range lowerItemSet {
		uniqueLowerDepthItems = append(uniqueLowerDepthItems, lowerItem)
	}
	return uniqueLowerDepthItems, nil
}
