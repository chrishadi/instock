package main

import "time"

type FilterResult struct {
	Active []Stock
	New    []Stock
	Stale  []Stock
}

func filter(newStocks []Stock, oldStocks []Stock) (FilterResult, error) {
	var res FilterResult

	if len(oldStocks) == 0 {
		res.Active = newStocks
		res.New = newStocks
		return res, nil
	}

	res.Active = make([]Stock, 0, len(newStocks))
	res.Stale = make([]Stock, 0)
	res.New = make([]Stock, 0)

	lastUpdates := make(map[string]string)
	for _, stock := range oldStocks {
		lastUpdates[stock.Code] = stock.LastUpdate
	}

	for _, stock := range newStocks {
		lastUpdate, exists := lastUpdates[stock.Code]
		if exists {
			last, err := time.Parse("2006-01-02 15:04:05", lastUpdate)
			if err != nil {
				return res, err
			}

			updatedAt, err := time.Parse("2006-01-02T15:04:05", stock.LastUpdate)
			if err != nil {
				return res, err
			}

			if !updatedAt.After(last) {
				res.Stale = append(res.Stale, stock)
				continue
			}
		} else {
			res.New = append(res.New, stock)
		}

		res.Active = append(res.Active, stock)
	}

	return res, nil
}
