package forex

import (
	"errors"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/joho/godotenv"
)

const TOLERANCE = 0.0001

func newService() (svc Service, err error) {
	godotenv.Load()
	apiKey := os.Getenv(`FOREX_API_KEY`)
	if apiKey == "" {
		err = errors.New(`An API key for Open Exchange Rates (https://openexchangerates.org) must be provided in the FOREX_API_KEY environment variable`)
		return
	}
	svc = Service{NewOpenExchangeRatesClient(apiKey, NewInMemRepository())}
	return
}
func TestConversion(t *testing.T) {
	svc, err := newService()
	if err != nil {
		t.Fatal(err)
	}
	tables := []struct {
		from                       string
		to                         string
		year                       int
		month                      int
		day                        int
		amount                     float32
		convertedAmountExpectation float32
		conversionRateExpectation  float32
	}{
		{`USD`, `EUR`, 2014, 12, 28, 1.0, 0.8211, 0.82105},
		{`EUR`, `USD`, 2014, 12, 28, 1.0, 1.2180, 1.2179526216430180865964313988186},
		{`EUR`, `EUR`, 2014, 12, 28, 1.0, 1.0, 1.0},
		{`GBP`, `JPY`, 2017, 8, 6, 1.0, 144.568069, 144.568069},
		{`GBP`, `JPY`, 2017, 8, 6, 10.0, 1445.68069, 144.568069},
	}
	for _, table := range tables {
		date := fmt.Sprintf("%d-%02d-%02d", table.year, table.month, table.day)
		convertedAmount, conversionRate, err := svc.Convert(table.from, table.to, table.year, table.month, table.day, table.amount)
		if err != nil {
			t.Errorf("unexpected error converting %02f %s to %s as of %s: %s", table.amount, table.from, table.to, date, err)
			continue
		}
		if !isCloseEnough(convertedAmount, table.convertedAmountExpectation, TOLERANCE) {
			t.Errorf("converting %02f %s => %s as of %s: expected %04f, got %04f", table.amount, table.from, table.to, date, table.convertedAmountExpectation, convertedAmount)
		}
		if !isCloseEnough(conversionRate, table.conversionRateExpectation, TOLERANCE) {
			t.Errorf("rate for %s => %s as of %s: expected %04f, got %04f", table.from, table.to, date, table.conversionRateExpectation, conversionRate)
		}
	}
}

func isCloseEnough(a, b, tolerance float32) bool {
	diff := math.Abs(float64(a) - float64(b))
	return diff < float64(tolerance)
}
