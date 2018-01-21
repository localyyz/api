package presenter

import (
	"fmt"
	"path"
	"strings"
)

func thumbImage(url string) string {
	if len(url) == 0 {
		return ""
	}
	imgURI, imgFile := path.Split(url)
	dotIdx := strings.LastIndex(imgFile, ".")
	return fmt.Sprintf("%s%s_medium.%s", imgURI, imgFile[:dotIdx], imgFile[dotIdx+1:])
}
