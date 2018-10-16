package sync

import (
	"strings"

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
		if strings.Contains(c.Value, "-") {
			c.IsSpecial = true
			c.Weight += 1
		}
		whitelistCache[c.Value] = append(whitelistCache[c.Value], *c)
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
