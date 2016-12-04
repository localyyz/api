package locale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"googlemaps.github.io/maps"

	"github.com/golang/geo/s2"
	"github.com/goware/lg"
)

// s2map api service

//curl 'http://s2map.com/api/s2cover'
// application/x-www-form-urlencoded
//'points=43.64773684556396%2C-79.40681934356688%2C43.65180470533265%2C-79.38737869262694%2C43.649941597649416%2C-79.38647747039795%2C43.649242917367836%2C-79.38965320587158%2C43.64890133738328%2C-79.39111232757567%2C43.64862186140564%2C-79.3924856185913%2C43.64815606521969%2C-79.39388036727904%2C43.64789211244429%2C-79.39525365829468%2C43.64772131885385%2C-79.39664840698241%2C43.6475971050285%2C-79.39767837524413%2C43.647348676607244%2C-79.39868688583373%2C43.646618912177715%2C-79.40237760543823%2C43.645827030107256%2C-79.40608978271484&min_level=1&max_level=30&max_cells=200&level_mod=1'

type S2Map struct {
	CellID       string         `json:"id"`
	CellIDSigned string         `json:"id_signed"`
	Token        string         `json:"token"`
	Pos          int64          `json:"pos"`
	Face         int32          `json:"face"`
	Level        int32          `json:"level"`
	LL           *maps.LatLng   `json:"ll"`
	Shape        []*maps.LatLng `json:"shapes"`
}

var (
	ApiUrl = "http://s2map.com/api/s2cover"
)

func GetS2Cover(coords []LatLng) (s2.CellUnion, error) {
	payload := url.Values{
		"min_level": {"1"},
		"max_level": {"30"},
		"max_cells": {"200"},
		"level_mod": {"1"},
	}

	strPoints := make([]string, len(coords))
	for i, p := range coords {
		strPoints[i] = fmt.Sprintf("%f,%f", p[1], p[0])
	}
	payload.Add("points", strings.Join(strPoints, ","))

	req, _ := http.NewRequest("POST", ApiUrl, bytes.NewBufferString(payload.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var s2maps []*S2Map
	if err := json.NewDecoder(resp.Body).Decode(&s2maps); err != nil {
		return nil, err
	}

	var cells s2.CellUnion
	for _, m := range s2maps {
		cid, _ := strconv.ParseUint(m.CellID, 10, 64)
		cells = append(cells, s2.CellID(cid))
	}

	lg.Infof("Found %d cells", len(cells))
	return cells, nil
}
