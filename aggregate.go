package ingest

import (
	"container/list"
	"time"
)

type AggregateResult struct {
	Active     []Stock
	New        []Stock
	Stale      []Stock
	TopGainers []string
	TopLosers  []string
}

func aggregate(newStocks []Stock, stockLastUpdates []StockLastUpdate, numOfGL int) (*AggregateResult, error) {
	var res AggregateResult

	if len(stockLastUpdates) == 0 {
		res.Active = newStocks
		res.New = newStocks
		return &res, nil
	}

	res.Active = make([]Stock, 0, len(newStocks))
	res.Stale = make([]Stock, 0)
	res.New = make([]Stock, 0)

	lastUpdateMap := make(map[string]string, len(stockLastUpdates))
	for _, stock := range stockLastUpdates {
		lastUpdateMap[stock.Code] = stock.LastUpdate
	}

	topGainers := list.New()
	topLosers := list.New()

	for _, stock := range newStocks {
		lastUpdate, exist := lastUpdateMap[stock.Code]
		if exist {
			last, err := time.Parse("2006-01-02 15:04:05", lastUpdate)
			if err != nil {
				return &res, err
			}

			updatedAt, err := time.Parse("2006-01-02T15:04:05", stock.LastUpdate)
			if err != nil {
				return &res, err
			}

			if !updatedAt.After(last) {
				res.Stale = append(res.Stale, stock)
				continue
			}
		} else {
			res.New = append(res.New, stock)
		}

		res.Active = append(res.Active, stock)

		gain := stock.OneDay
		if gain == 0.0 {
			continue
		}
		if gain > 0.0 {
			if topGainers.Len() < numOfGL || gain > topGainers.Back().Value.(StockGain).Gain {
				updateTopRank(topGainers, stock, func(a, b float64) bool { return a > b }, numOfGL)
			}
		} else {
			if topLosers.Len() < numOfGL || gain < topLosers.Back().Value.(StockGain).Gain {
				updateTopRank(topLosers, stock, func(a, b float64) bool { return a < b }, numOfGL)
			}
		}
	}

	res.TopGainers = extractTopRankCodes(topGainers)
	res.TopLosers = extractTopRankCodes(topLosers)

	return &res, nil
}

func updateTopRank(tr *list.List, stock Stock, greater func(a, b float64) bool, limit int) {
	gain := stock.OneDay
	obj := StockGain{stock.Code, gain}

	if tr.Len() == 0 || greater(gain, tr.Front().Value.(StockGain).Gain) {
		tr.PushFront(obj)
	} else {
		e := tr.Back()
		for greater(gain, e.Value.(StockGain).Gain) {
			e = e.Prev()
		}
		tr.InsertAfter(obj, e)
	}

	if tr.Len() > limit {
		tr.Remove(tr.Back())
	}
}

func extractTopRankCodes(ls *list.List) []string {
	codes := make([]string, 0, ls.Len())
	for e := ls.Front(); e != nil; e = e.Next() {
		codes = append(codes, e.Value.(StockGain).Code)
	}
	return codes
}
