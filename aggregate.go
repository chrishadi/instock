package ingest

import (
	"container/list"
	"time"

	"github.com/chrishadi/instock/toplist"
)

type AggregateResult struct {
	Active     []Stock
	New        []Stock
	Stale      []Stock
	TopGainers []string
	TopLosers  []string
}

type StockGain struct {
	Code string
	Gain float64
}

func moreGain(a, b interface{}) bool { return a.(StockGain).Gain > b.(StockGain).Gain }
func moreLoss(a, b interface{}) bool { return a.(StockGain).Gain < b.(StockGain).Gain }

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

	topGainers := toplist.New(numOfGL, moreGain)
	topLosers := toplist.New(numOfGL, moreLoss)

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
			topGainers.Add(StockGain{stock.Code, stock.OneDay})
		} else {
			topLosers.Add(StockGain{stock.Code, stock.OneDay})
		}
	}

	res.TopGainers = extractTopRankCodes(topGainers.Elements())
	res.TopLosers = extractTopRankCodes(topLosers.Elements())

	return &res, nil
}

func extractTopRankCodes(ls *list.List) []string {
	codes := make([]string, 0, ls.Len())
	for e := ls.Front(); e != nil; e = e.Next() {
		codes = append(codes, e.Value.(StockGain).Code)
	}
	return codes
}
