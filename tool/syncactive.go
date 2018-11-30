package tool

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/lib/connect"
	"bitbucket.org/moodie-app/moodie-api/lib/shopify"
	"github.com/go-chi/render"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
)

type wrapper struct {
	ID     int64         `json:"id"`
	ErrMsg string        `json:"errMsg"`
	Shop   *shopify.Shop `json:"shop"`

	place  *data.Place
	err    error
	bucket <-chan time.Time
}

func gen(done <-chan struct{}, places ...*data.Place) <-chan *wrapper {
	bucket := throttleBucket(done)
	out := make(chan *wrapper)
	go func() {
		defer close(out)
		for i, p := range places {
			select {
			case out <- &wrapper{place: p, bucket: bucket}:
			case <-done:
				// early exit
				return
			}
			if i > 0 && i%400 == 0 {
				lg.Infof("gen %d", i)
			}
		}
	}()
	return out
}

func doHead(done <-chan struct{}, in <-chan *wrapper) <-chan *wrapper {
	out := make(chan *wrapper)
	go func() {
		defer close(out)
		for w := range in {

			u := fmt.Sprintf("https://%s.myshopify.com", w.place.ShopifyID)
			resp, err := http.Head(u)
			if err != nil {
				w.err = err
			} else {
				if resp.StatusCode == http.StatusNotFound {
					w.err = errors.New("not found")
				} else if resp.StatusCode == http.StatusPaymentRequired {
					w.err = errors.New("payment required")
				} else if resp.StatusCode != http.StatusOK {
					w.err = fmt.Errorf("http status: %v", resp.Status)
				}
			}
			select {
			case out <- w:
			case <-done:
				// early exit.
				return
			}
		}
	}()
	return out
}

func doApi(done <-chan struct{}, in <-chan *wrapper) <-chan *wrapper {
	out := make(chan *wrapper)
	go func() {
		defer close(out)
		for w := range in {
			do := func() (*shopify.Shop, error) {
				// make sure previous step did not return error
				if w.err != nil {
					return nil, w.err
				}

				// block if we're out of tokens
				<-w.bucket

				//fetch api
				client, err := connect.GetShopifyClient(w.place.ID)
				if err != nil {
					return nil, err
				}

				shop, resp, err := client.Shop.Get(context.Background())
				if err != nil {
					return nil, err
				}
				if resp.StatusCode != http.StatusOK {
					// unauthorized. it means the merchant uninstalled us
					if resp.StatusCode == http.StatusUnauthorized {
						return nil, ErrUnauthorized
					}
					if resp.StatusCode == http.StatusTooManyRequests {
						// this shouldn't be.. our ticker should have handled
						// it. just echo and move on.
						return nil, nil
					}
					return nil, fmt.Errorf(resp.Status)
				}

				return shop, nil
			}
			w.Shop, w.err = do()

			select {
			case out <- w:
			case <-done:
				// early exit.
				return
			}
		}
	}()

	return out
}

func merge(done <-chan struct{}, cs ...<-chan *wrapper) <-chan *wrapper {
	var wg sync.WaitGroup
	out := make(chan *wrapper)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan *wrapper) {
		defer wg.Done()
		for w := range c {
			// lift up to the top
			w.ID = w.place.ID
			if w.err != nil {
				w.ErrMsg = w.err.Error()
			}
			select {
			case out <- w:
			case <-done:
				// early exit.
				return
			}
		}
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func IsClosed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

func throttleBucket(done <-chan struct{}) <-chan time.Time {
	// bucket throttler
	rate := time.Second / 2
	bucket := make(chan time.Time, 2)

	tick := time.NewTicker(rate)
	go func() {
		defer tick.Stop()
		for t := range tick.C {
			select {
			case bucket <- t:
			case <-done:
				return
			default:
				// block
			}
		}
	}()
	return bucket
}

func activeMerchantSync(w http.ResponseWriter, r *http.Request) {
	cond := db.Cond{}
	var places []*data.Place
	data.DB.Place.Find(cond).OrderBy("id").All(&places)
	lg.Infof("processing %d merchants", len(places))

	// interupt and cancel
	done := make(chan struct{})

	// generate inputs
	in := gen(done, places...)

	var workers []<-chan *wrapper
	for i := 0; i < 10; i++ {
		workers = append(workers, doApi(done, in))
	}

	var uninstalled []int64
	list := []*wrapper{}
	for w := range merge(done, workers...) {
		if w.err != nil {
			lg.Warnf("place(%s,%d) - other error %v", w.place.Name, w.place.ID, w.err)
			uninstalled = append(uninstalled, w.place.ID)
			continue
		}
		list = append(list, w)
	}

	// TODO: Use request context to handle cancelation
	if !IsClosed(done) {
		close(done)
	}
	lg.Infof("uinstalled(%d) %v", len(uninstalled), uninstalled)
	render.Respond(w, r, list)
}
