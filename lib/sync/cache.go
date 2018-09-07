package sync

import (
	"bitbucket.org/moodie-app/moodie-api/data"
	"github.com/pkg/errors"
	"github.com/pressly/lg"
)

type (
	whitelist map[string][]data.Whitelist
	blacklist map[string]data.Blacklist
)

var (
	whitelistCache whitelist
	blacklistCache blacklist

	tagSpecial = []string{
		"bomber-jacket",
		"boxer-brief",
		"boxer-trunk",
		"eau-de-toilette",
		"eau-de-parfum",
		"eau-de-cologne",
		"face-mask",
		"after-shave",
		"face-wash",
		"beard-kit",
		"face-cream",
		"night-cream",
		"cleansing-gel",
		"bath-bomb",
		"nail-polish",
		"lip-gloss",
		"body-cream",
		"body-wash",
		"bath-salt",
		"bath-oil",
		"body-butter",
		"eye-liner",
		"eye-shadow",
		"foot-cream",
		"toe-separator",
		"pocket-square",
		"shoulder-bag",
		"sleeping-bag",
		"t-shirt",
		"track-pant",
		"v-neck",
		"air-jordan",
		"lace-up",
		"slip-on",
		"lexi-noel",
		"sport-coat",
	}
)

func SetupWhitelistCache() error {
	keywords, err := data.DB.Whitelist.FindAll(nil)
	if err != nil {
		return errors.Wrap(err, "whitelist cache")
	}

	// there *may* be similar value whitelisted terms
	// for both male and female gender
	whitelistCache = make(map[string][]data.Whitelist)
	for _, c := range keywords {
		whitelistCache[c.Value] = append(whitelistCache[c.Value], *c)
	}

	// load up predefined special tag list (not in db)
	// TODO: migrate to db
	for _, t := range tagSpecial {
		w, ok := whitelistCache[t]
		if !ok {
			w = []data.Whitelist{
				{Value: t, Gender: data.ProductGenderUnisex, IsSpecial: true},
			}
		}
		w[0].IsSpecial = true
	}
	lg.Infof("whitelist cache: keys(%d)", len(whitelistCache))

	return nil
}

func SetupBlacklistCache() error {
	keywords, err := data.DB.Blacklist.FindAll(nil)
	if err != nil {
		return errors.Wrap(err, "blacklist cache")
	}

	blacklistCache := make(map[string]data.Blacklist)
	for _, word := range keywords {
		blacklistCache[word.Word] = *word
	}
	lg.Infof("blacklist cache: keys(%d)", len(blacklistCache))

	return nil
}

func SetupCache() error {
	if data.DB == nil {
		return data.ErrUnconfigured
	}
	if err := SetupWhitelistCache(); err != nil {
		return err
	}
	if err := SetupBlacklistCache(); err != nil {
		return err
	}
	return nil
}
