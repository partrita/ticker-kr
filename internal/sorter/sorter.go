package sorter

import (
	"sort"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// Sorter represents a function that sorts quotes
type Sorter func([]*c.Asset) []*c.Asset

// NewSorter creates a sorting function
func NewSorter(sort string) Sorter {
	var sortDict = map[string]Sorter{
		"alpha": sortByAlpha,
		"value": sortByValue,
		"user":  sortByUser,
	}
	if sorter, ok := sortDict[sort]; ok {
		return sorter
	}

	return sortByChange
}

func sortByUser(assets []*c.Asset) []*c.Asset {

	assetCount := len(assets)

	if assetCount <= 0 {
		return assets
	}

	sort.SliceStable(assets, func(i, j int) bool {
		return assets[j].Meta.OrderIndex > assets[i].Meta.OrderIndex
	})

	return assets

}

func sortByAlpha(assets []*c.Asset) []*c.Asset {

	assetCount := len(assets)

	if assetCount <= 0 {
		return assets
	}

	sort.SliceStable(assets, func(i, j int) bool {
		return assets[j].Symbol > assets[i].Symbol
	})

	return assets
}

func sortByValue(assets []*c.Asset) []*c.Asset {

	assetCount := len(assets)

	if assetCount <= 0 {
		return assets
	}

	sort.SliceStable(assets, func(i, j int) bool {
		if assets[i].Exchange.IsActive != assets[j].Exchange.IsActive {
			return assets[i].Exchange.IsActive
		}
		return assets[i].Position.Value > assets[j].Position.Value
	})

	return assets
}

func sortByChange(assets []*c.Asset) []*c.Asset {

	assetCount := len(assets)

	if assetCount <= 0 {
		return assets
	}

	sort.SliceStable(assets, func(i, j int) bool {
		if assets[i].Exchange.IsActive != assets[j].Exchange.IsActive {
			return assets[i].Exchange.IsActive
		}
		return assets[i].QuotePrice.ChangePercent > assets[j].QuotePrice.ChangePercent
	})

	return assets

}
