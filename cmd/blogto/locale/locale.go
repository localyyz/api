package locale

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sort"
)

type Locale struct {
	ID          int64    `json:"id"`
	Boundaries  []LatLng `json:"boundaries"`
	Description string   `json:"description"`
	Name        string   `json:"name"`
	Shorthand   string   `json:"shorthand"`
}
type Locales []*Locale

var (
	cache     Locales
	LocaleMap map[string]*Locale
)

func (slice Locales) Len() int {
	return len(slice)
}

func (slice Locales) Less(i, j int) bool {
	return slice[i].ID < slice[j].ID
}

func (slice Locales) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func List() {
	localeList := ""
	for _, loc := range cache {
		localeList += fmt.Sprintf(" %d", loc.ID)
		localeList += fmt.Sprintf(" %s", loc.Name)
		localeList += fmt.Sprintf(" (%s) \n", loc.Shorthand)
	}
	fmt.Println(localeList)
}

func loadData() {
	//loc := locales[*localeID]

	//var dbLocale *data.Locale
	//err := data.DB.Locale.Find(db.Cond{"shorthand": loc.Shorthand}).One(&dbLocale)
	//if err != nil {
	//if err != db.ErrNoMoreRows {
	//log.Fatal(err)
	//}
	//desc, _ := url.QueryUnescape(loc.Description)
	//dbLocale = &data.Locale{
	//Shorthand:   loc.Shorthand,
	//Name:        loc.Name,
	//Description: desc,
	//}
	//if err := data.DB.Locale.Save(dbLocale); err != nil {
	//log.Fatal(err)
	//}
	//}
	//lg.Warnf("found locale(%v): %s", dbLocale.ID, dbLocale.Name)

	//for _, c := range getCells(&loc) {
	//cell := &data.Cell{LocaleID: dbLocale.ID, CellID: int64(c)}
	//if err := data.DB.Cell.Save(cell); err != nil {
	//lg.Warn(errors.Wrapf(err, "failed inserting cell(%v)", int64(c)))
	//break
	//}
	//}
}

func init() {
	f, err := os.Open("cmd/blogto/data/locale.json")
	if err != nil {
		log.Fatal(err)
	}

	buf, _ := ioutil.ReadAll(f)
	if err := json.Unmarshal(buf, &cache); err != nil {
		log.Fatal(err)
	}
	sort.Sort(cache)

	// write to map for easy lookup
	LocaleMap = make(map[string]*Locale, len(cache))
	for _, l := range cache {
		LocaleMap[l.Shorthand] = l
	}
}
