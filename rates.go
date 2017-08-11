package forex

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type EndOfDay struct {
	Timestamp int                `json:"timestamp"` // seconds since the UNIX epoch
	Base      string             `json:"base"`      // ISO 4217 base currency code
	Rates     map[string]float32 `json:"rates"`     // map of ISO 4217 currency codes to exchange rates
}

type DataSource interface {
	GetByDate(year, month, day int) (EndOfDay, error)
}

type openExchangeRatesService struct {
	appID     string
	repo      Repository
	locks     map[string]*sync.Mutex
	locksLock *sync.Mutex
}

type Repository interface {
	Get(year, month, day int) (EndOfDay, bool)
	Set(year int, month int, day int, data EndOfDay) error
}

func (s openExchangeRatesService) GetByDate(year, month, day int) (eod EndOfDay, err error) {
	dstr := fmt.Sprintf("%d-%02d-%02d", year, month, day)
	s.locksLock.Lock()
	_, ok := s.locks[dstr]
	if !ok {
		s.locks[dstr] = &sync.Mutex{}
	}
	s.locksLock.Unlock()
	s.locks[dstr].Lock()
	defer s.locks[dstr].Unlock()
	eod, ok = s.repo.Get(year, month, day)
	if ok {
		return eod, nil
	}
	resp, httpErr := http.Get(fmt.Sprintf(`https://openexchangerates.org/api/historical/%s.json?app_id=%s`, dstr, s.appID))
	if httpErr != nil {
		return eod, httpErr
	}
	defer resp.Body.Close()
	dec := json.NewDecoder(resp.Body)
	err = dec.Decode(&eod)
	if err != nil {
		err = s.repo.Set(year, month, day, eod)
	}
	return
}

func NewOpenExchangeRatesClient(appID string, repo Repository) DataSource {
	return openExchangeRatesService{
		appID:     appID,
		repo:      repo,
		locks:     make(map[string]*sync.Mutex),
		locksLock: new(sync.Mutex),
	}
}

type inMemRepo struct {
	data map[string]EndOfDay
	*sync.RWMutex
}

func (r inMemRepo) Get(year, month, day int) (eod EndOfDay, ok bool) {
	r.RLock()
	eod, ok = r.data[fmt.Sprintf("%d-%02d-%02d", year, month, day)]
	r.RUnlock()
	return
}

func (r inMemRepo) Set(year, month, day int, eod EndOfDay) error {
	r.Lock()
	r.data[fmt.Sprintf("%d-%02d-%02d", year, month, day)] = eod
	r.Unlock()
	return nil
}

func NewInMemRepository() Repository {
	return inMemRepo{
		data:    make(map[string]EndOfDay),
		RWMutex: new(sync.RWMutex),
	}
}
