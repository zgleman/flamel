package cors

import (
	"fmt"
	"net/http"
)

type Cors struct {
	headers string
	methods string
	origins []string
	//seconds to cache the response
	MaxAgeSeconds int
	//Accelerated Mobile Page support
	amp      bool
	ampFetch map[string]bool
}

const ampAllowedOriginAmpproject string = ".ampproject.org"
const ampAllowedOriginCloudflare string = ".amp.cloudflare.com"

const KeyAmpSourceOrigin string = "__amp_source_origin"
const KeyAmpSameOriginHeader string = "AMP-Same-Origin"

func NewCors(origins []string, methods []string, headers []string) *Cors {

	c := Cors{}
	c.amp = false
	c.headers = convertToHeaderString(headers)
	c.methods = convertToHeaderString(methods)
	c.origins = origins

	return &c
}

//sets the list of allowed AMP urls
func (c *Cors) EnableAmpFetch(urls []string) {
	c.ampFetch = make(map[string]bool, len(urls))
	for _, v := range urls {
		c.ampFetch[v] = true
	}
	c.amp = true
}

func (c Cors) AMP() bool {
	return c.amp
}

func (c Cors) AMPForUrl(url string) bool {
	if !c.AMP() {
		return false
	}

	_, valid := c.ampFetch[url]
	return valid
}

//returns true if origin has been allowed
func (c *Cors) HandleOptions(w http.ResponseWriter, origin string) bool {
	allowed := false
	/*if we support AMP we check only for:
	* 1. *.ampproject.org
	* 2. *.amp.cloudflare.com
	* 3. our origin
	* for reference: https://github.com/ampproject/amphtml/blob/master/spec/amp-cors-requests.md
	 */
	if c.amp && (origin[len(origin)-len(ampAllowedOriginAmpproject):] == ampAllowedOriginAmpproject ||
		origin[len(origin)-len(ampAllowedOriginCloudflare):] == ampAllowedOriginCloudflare) {

		allowed = true
		w.Header().Set("Access-Control-Allow-Origin", origin)

	} else {
		//process the allowed origins, including the case 3 (our origin)
		for _, v := range c.origins {
			if v == origin {
				allowed = true
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
	}

	if c.methods != "" {
		w.Header().Set("Access-Control-Allow-Methods", c.methods)
	}

	if c.headers != "" {
		w.Header().Set("Access-Control-Allow-Headers", c.headers)
	}

	if c.MaxAgeSeconds > 0 {
		age := fmt.Sprintf("%d", c.MaxAgeSeconds)
		w.Header().Set("Access-Control-Max-Age", age)
	}

	return allowed
}

func (c *Cors) ValidateAMP(w http.ResponseWriter, source string) error {
	for _, v := range c.origins {
		if v == source {
			w.Header().Set("AMP-Access-Control-Allow-Source-Origin", source)
			return nil
		}
	}
	return fmt.Errorf("invalid AMP origin request! Source is: %s", source)
}

func convertToHeaderString(values []string) string {

	s := ""
	for i, v := range values {
		if i == 0 {
			s = fmt.Sprintf("%s", v)
			continue
		}

		s = fmt.Sprintf("%s, %s", s, v)
	}

	return s
}
