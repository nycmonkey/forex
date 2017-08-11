package forex

import "errors"

// Convert applies the end-of-day foreign exchange rate for a specified currency pair, returning the converted amount
// and the rate used.  The "from" and "to" currencies must be specified using ISO 4217 currency codes, and the as-of date
// must be specified using the format YYYY-MM-DD
func (svc Service) Convert(from string, to string, year int, month int, day int, amount float32) (convertedAmount float32, rate float32, err error) {
	if from == to {
		return amount, 1.0, nil
	}
	if year < 2000 {
		err = errors.New("Data not available before the year 2000")
		return
	}
	if month < 1 || month > 12 {
		err = errors.New("Invalid month")
		return
	}
	if day < 1 || day > 31 {
		err = errors.New("Invalid day")
		return
	}
	eod, err := svc.GetByDate(year, month, day)
	if err != nil {
		return
	}
	if from == `USD` {
		var ok bool
		rate, ok = eod.Rates[to]
		if !ok {
			err = errors.New("Unrecognized currency code: " + to)
			return
		}
		convertedAmount = amount * rate
		return
	}
	if to == `USD` {
		invRate, ok := eod.Rates[from]
		if !ok {
			err = errors.New("Unrecognized currency code: " + from)
			return
		}
		if invRate == 0 {
			err = errors.New("Rate not available")
			return
		}
		rate = 1.0 / invRate
		convertedAmount = amount * rate
		return
	}
	// convert to USD, then to target currency
	invRate, ok := eod.Rates[from]
	if !ok {
		err = errors.New("Unrecognized currency code: " + from)
		return
	}
	if invRate == 0 {
		err = errors.New("Rate not available")
		return
	}
	r1 := 1.0 / invRate
	var r2 float32
	r2, ok = eod.Rates[to]
	if !ok {
		err = errors.New("Unrecognized currency code: " + to)
		return
	}
	rate = r1 * r2
	convertedAmount = amount * rate
	return
}

type Service struct {
	DataSource
}
