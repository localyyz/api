package search

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"

	"bitbucket.org/moodie-app/moodie-api/data"
	"bitbucket.org/moodie-app/moodie-api/web/api"
	"github.com/go-chi/render"
	"github.com/pressly/lg"
	db "upper.io/db.v3"
)

type RelatedWord struct {
	Lexeme    string `json:"lexeme"`
	Positions []int  `json:"positions"`
}

func isGender(word string) bool {
	switch word {
	case "woman", "women", "female", "man", "men", "male":
		return true
	default:
		return false
	}
}

func RelatedTags(w http.ResponseWriter, r *http.Request) {
	var p omniSearchRequest
	if err := render.Bind(r, &p); err != nil {
		render.Render(w, r, api.ErrInvalidRequest(err))
		return
	}

	genderQuery := db.Raw("1 = 1")
	if gender, ok := r.Context().Value("session.gender").(data.UserGender); ok {
		genderQuery = db.Raw(fmt.Sprintf("p.gender = %d", gender))
	}

	filterTags := append(p.rawParts, p.FilterTags...)
	filterTags = append(filterTags, p.queryParts...)

	rows, err := data.DB.Select(db.Raw(`json_agg(json_build_object('lexeme', lexeme, 'positions', positions)) as words`)).
		From(db.Raw(`(
			SELECT p.id, (unnest(to_tsvector('simple', title))).*
			FROM products p
			LEFT JOIN places pl ON pl.id = p.place_id
			WHERE tsv @@ phraseto_tsquery(?)
			AND pl.status = 3
			AND p.deleted_at IS NULL
			AND pl.weight > 4
			AND p.category != '{}'
			AND ?
			LIMIT 1000
		) x`, p.Query, genderQuery)).
		Where(db.Cond{
			db.Raw("to_tsvector(lexeme)"): db.NotEq(""),
			"lexeme":                      db.NotIn(filterTags),
		}).
		GroupBy("id").
		Query()
	if err != nil {
		render.Respond(w, r, err)
		return
	}
	defer rows.Close()

	wordMap := map[string]int{}
	for rows.Next() {
		var rawWords []byte
		if err := rows.Scan(&rawWords); err != nil {
			lg.Warn(err)
			break
		}

		var words []RelatedWord
		if err := json.Unmarshal(rawWords, &words); err != nil {
			lg.Warn(err)
			break
		}

		wordPositions := map[int]string{}
		for _, w := range words {
			if isGender(w.Lexeme) || len(w.Lexeme) < 2 {
				continue
			}
			for _, p := range w.Positions {
				wordPositions[p] = w.Lexeme
			}
		}

		// iterate over words and aggregate groupings
		for j := 1; j <= len(wordPositions); j++ {
			if wordPositions[j] == "" {
				continue
			}
			if j2 := j + 1; j2 < len(wordPositions) {
				if wordPositions[j2] == "" {
					continue
				}

				w := fmt.Sprintf("%s %s", wordPositions[j], wordPositions[j2])
				ww := fmt.Sprintf("%s %s", wordPositions[j2], wordPositions[j])
				if _, ok := wordMap[ww]; ok && wordMap[ww] > wordMap[w] {
					wordMap[ww]++
					continue
				}
				wordMap[w]++
			}
		}
	}
	if err := rows.Err(); err != nil {
		render.Respond(w, r, err)
		return
	}

	// sorted map
	type kv struct {
		Key   string
		Value int
	}
	var ss []kv
	for k, v := range wordMap {
		ss = append(ss, kv{k, v})
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Value > ss[j].Value
	})
	var relatedTags []string
	for i, kv := range ss {
		if i == 10 {
			break
		}
		relatedTags = append(relatedTags, kv.Key)
	}

	render.Respond(w, r, relatedTags)
}
