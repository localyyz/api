package htmlx

import (
	"bytes"
	h "html"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/microcosm-cc/bluemonday"
	gnHtml "golang.org/x/net/html"
	gnAtom "golang.org/x/net/html/atom"
)

var (
	bodyCaptionizeParsePolicy *bluemonday.Policy
)

var (
	reCaptionizeHtmlWithPolicy = regexp.MustCompile(`(\w)(<.+>)?<\/p><p>(<.+>)?(\w)`)
	// Regex from RegexBuddy	http://stackoverflow.com/questions/161738/what-is-the-best-regular-expression-to-check-if-a-string-is-a-valid-url/190405#190405
	reConvertLinksToAnchorTags = regexp.MustCompile(`(\s|^)((https?|ftp|file):\/\/?[-A-Za-z0-9+&@#\/%?=~_|!:,.;]+[-A-Za-z0-9+&@#\/%=~_|])(\s|$)`)
)

func init() {
	//bodyCaptionizeParsePolicy allows images, etc
	bodyCaptionizeParsePolicy = bluemonday.NewPolicy()
	bodyCaptionizeParsePolicy.AllowStandardURLs()
	bodyCaptionizeParsePolicy.RequireNoFollowOnLinks(false)
	bodyCaptionizeParsePolicy.RequireParseableURLs(true)
	bodyCaptionizeParsePolicy.AllowRelativeURLs(false)
	bodyCaptionizeParsePolicy.AddTargetBlankToFullyQualifiedLinks(true)
	bodyCaptionizeParsePolicy.AllowAttrs("href").OnElements("a")
	bodyCaptionizeParsePolicy.AllowAttrs("src").OnElements("img")
	bodyCaptionizeParsePolicy.AllowElements("b", "strong", "i", "em", "u", "br", "p", "h1", "h2", "h3", "h4", "h5", "h6", "hr", "li", "ul", "ol", "blockquote", "pre", "code")

}

func CaptionizeHtmlBody(inHTML string, maxLen int) string {
	return RemoveNewlines(CaptionizeHtmlWithPolicy(bodyCaptionizeParsePolicy, inHTML, maxLen))
}

func RemoveNewlines(s string) string {
	return strings.Replace(strings.Replace(strings.Replace(strings.Replace(s, "\r\n", " ", -1), "\n", " ", -1), "  ", " ", -1), "\\n", " ", -1)
}

func TagifyNewlines(s string) string {
	return strings.Replace(strings.Replace(s, "\r\n", "<br />", -1), "\n", "<br />", -1)
}

// maxLen can be -1, which means do not truncate
func CaptionizeHtmlWithPolicy(p *bluemonday.Policy, inHTML string, maxLen int) string {
	// Regex based on https://github.com/pressly/pressly-data/blob/0968cb37cc9cd9a9984209b093cc12b77a35c930/lib/pressly-data/util/html.rb#L27
	hypenateHtmlElements := func(in string) string {
		out := reCaptionizeHtmlWithPolicy.ReplaceAllString(in, "$1$2</p> <p>$3$4")
		return out
	}

	convertLinksToAnchorTags := func(in string) string {
		out := reConvertLinksToAnchorTags.ReplaceAllString(in, " <a href=\"$2\" target=\"_blank\">$2</a> ")
		return out
	}

	// should unescape any escaped html
	// or tags won't be filtered correctly
	outHTML := h.UnescapeString(inHTML)

	outHTML = hypenateHtmlElements(outHTML)
	outHTML = convertLinksToAnchorTags(outHTML)
	outHTML = strings.Replace(outHTML, "\t", " ", -1)

	// NOTE/HACK: Do this until either outHTML is empty or not changed..
	// This is to prevent nested inline tags to pass unnoticed.
	// TODO: Benchmark
	tmpHTML := outHTML
	for {
		outHTML = h.UnescapeString(p.Sanitize(outHTML))
		if len(outHTML) == 0 || tmpHTML == outHTML {
			break
		}
		tmpHTML = outHTML
	}
	outHTML = RemoveEmptyAnchorNodes(outHTML)

	if maxLen > 0 {
		outHTML = TruncateHtml(outHTML, maxLen)
	}

	outHTML = strings.Replace(outHTML, "  ", " ", -1)
	outHTML = strings.TrimSpace(outHTML)

	return h.UnescapeString(outHTML)
}

// This function removes tags such as <a href="http://abc.com">_</a>
//  where no text is present for the anchor link
func RemoveEmptyAnchorNodes(s string) string {
	nodes, err := gnHtml.ParseFragment(strings.NewReader(s), &gnHtml.Node{
		Type:     gnHtml.ElementNode,
		Data:     "body",
		DataAtom: gnAtom.Body,
	})
	if err != nil {
		return s
	}

	IsEmptyAnchorNodeFunc := func(node *gnHtml.Node) bool {
		if node.Type == gnHtml.ElementNode && node.Data == "a" && node.FirstChild == nil && node.LastChild == nil {
			// This mean this node itself is an empty anchor node
			return true
		}
		return false
	}

	var processNode func(*gnHtml.Node) bool
	processNode = func(n *gnHtml.Node) bool {
		// Returns True if this node is an empty anchor node. If not, it removes any empty anchor
		// nodes amongst its children and returns False. Also calls the function recursively
		// for each of its children.
		if IsEmptyAnchorNodeFunc(n) {
			return true
		}

		emptyAnchorNodes := make([]*gnHtml.Node, 0)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if IsEmptyAnchorNodeFunc(c) {
				emptyAnchorNodes = append(emptyAnchorNodes, c)
			} else {
				processNode(c)
			}
		}
		for _, emptyNode := range emptyAnchorNodes {
			n.RemoveChild(emptyNode)
		}
		return false
	}

	var w bytes.Buffer
	for _, node := range nodes {
		if !IsEmptyAnchorNodeFunc(node) {
			err := gnHtml.Render(&w, node)
			if err != nil {
				return s
			}
		}
	}
	return w.String()
}

func TruncateHtml(s string, maxLen int) string {
	// The text can be plain text or html snippet
	nodes, err := gnHtml.ParseFragment(strings.NewReader(s), &gnHtml.Node{
		Type:     gnHtml.ElementNode,
		Data:     "body",
		DataAtom: gnAtom.Body,
	})
	if err != nil {
		return s
	}

	keepGoing := true
	lengthSoFar := 0

	var f func(n *gnHtml.Node)
	f = func(n *gnHtml.Node) {

		if n.Type == gnHtml.TextNode {
			if keepGoing == true {
				utf8len := utf8.RuneCountInString(n.Data)
				if utf8len+lengthSoFar >= maxLen {

					x := maxLen - lengthSoFar - 3
					if x < 0 {
						x = 0

					} else if x > utf8len {
						x = utf8len
					}

					n.Data = unicodeSlice(n.Data, 0, x)
					n.Data += "..."
					keepGoing = false
				}
				lengthSoFar += utf8.RuneCountInString(n.Data)
			}
		}

		nodesToRemove := make([]*gnHtml.Node, 0)
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if keepGoing == false {
				nodesToRemove = append(nodesToRemove, c)
			} else {
				f(c)
			}
		}
		for _, node := range nodesToRemove {
			n.RemoveChild(node)
		}
	}

	nodesToRender := make([]*gnHtml.Node, 0)
	for _, node := range nodes {
		if keepGoing {
			f(node)
			nodesToRender = append(nodesToRender, node)
		}
	}

	var w bytes.Buffer
	for _, node := range nodesToRender {
		err := gnHtml.Render(&w, node)
		if err != nil {
			return s
		}
	}
	return w.String()
}

func unicodeSlice(s string, i int, j int) string {
	utf8len := utf8.RuneCountInString(s)
	if utf8len <= i {
		return ""
	}

	utf8 := []rune(s)
	if utf8len <= j {
		return string(utf8[i:utf8len])
	}

	return string(utf8[i:j])
}
