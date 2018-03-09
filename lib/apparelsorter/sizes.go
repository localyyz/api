package apparelsorter

import (
	"fmt"
	"regexp"
	"strconv"
)

const postpendIdx = int(^uint(0) >> 1)
const shoesizes = `([0-9]+)(\.([0-9]+))?`

var apparelsizes = []string{
	"^osfa.*$",
	"^one .*$",
	"^one$",
	"^xxs",
	"^xs .*$",
	"^x sm.*$",
	"^xs.*$",
	"^.* xs$",
	"^xs",
	"^extra small$",
	"^sm.*$",
	"^.* small",
	"^ss",
	"^short sleeve",
	"^ls",
	"^long sleeve",
	"^s$",
	"^small.*$",
	`^s\/.*$`,
	`^s \/.*$`,
	"^s .*$",
	"^m$",
	"^medium.*$",
	"^.*med.*$",
	"^m .*$",
	"^m[A-Za-z]*",
	`^M\/.*$`,
	"^l$",
	"^.*lg.*$",
	"^large.*$",
	"^l .*$",
	`^l\/.*$`,
	"^lt$",
	"^xl.*$",
	"^extra large$",
	"^x large.*$",
	"^.* XL$",
	"^x-l.*$",
	"^l[A-Za-z]*$",
	"^petite l.*$",
	"^1x.*$",
	"^.* 1x$",
	"^2x.*$",
	"^.* 2X$",
	"^.*XXL.*$",
	"^3x.*$",
	"^4x.*$",
	"^5x.*$",
	"^6x.*$",
	"^7x.*$",
	"^8x.*$",
	"^9x.*$",
	"^10x.*$",
	"^11x.*$",
	"^12x.*$",
	"^13x.*$",
	"^14x.*$",
	"^15x.*$",
	"^16x.*$",
	"^17x.*$",
	"^18x.*$",
}

var (
	apparelRegexes []*regexp.Regexp
	shoeRegex      *regexp.Regexp
)

func matchSize(s string) *Size {
	for i, re := range apparelRegexes {
		if re.MatchString(s) {
			return &Size{Size: s, Order: i}
		}
	}
	if part := shoeRegex.FindString(s); len(part) > 0 {
		// check if integer or float
		s := &Size{Size: s}
		if o, err := strconv.Atoi(part); err == nil {
			// Store as normalized integer
			s.Order = o * 1000
		} else {
			// Store as normalize float converted to integer
			if o, err := strconv.ParseFloat(part, 64); err == nil {
				s.Order = int(o * 1000.0)
			}
		}
		return s
	}
	// unmatched size. postpend
	return &Size{Size: s, Order: postpendIdx}
}

func init() {
	for _, rx := range apparelsizes {
		ri := fmt.Sprintf("(?i)%s", rx)
		re := regexp.MustCompile(ri)
		apparelRegexes = append(apparelRegexes, re)
	}
	shoeRegex = regexp.MustCompile(shoesizes)
}
