package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/mattn/go-colorable"
	logger "github.com/sirupsen/logrus"
	"html/template"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type nineGagEntity struct {
	Status string `json:"status"`
	Items  []struct {
		SrcAlias     string `json:"src_alias,omitempty"`
		ID           string `json:"id,omitempty"`
		URL          string `json:"url,omitempty"`
		Title        string `json:"title,omitempty"`
		Type         string `json:"type,omitempty"`
		CreationTs   int    `json:"creationTs,omitempty"`
		UpVote       int    `json:"upVote,omitempty"`
		DownVote     int    `json:"downVote,omitempty"`
		UrlsResource struct {
			Image700 struct {
				Width  int    `json:"width"`
				Height int    `json:"height"`
				URL    string `json:"url"`
			} `json:"image700"`
			Image460 struct {
				Width   int    `json:"width"`
				Height  int    `json:"height"`
				URL     string `json:"url"`
				WebpURL string `json:"webpUrl"`
			} `json:"image460"`
			ImageFbThumbnail struct {
				Width  int    `json:"width"`
				Height int    `json:"height"`
				URL    string `json:"url"`
			} `json:"imageFbThumbnail"`
			Image460Sv struct {
				Width    int    `json:"width"`
				Height   int    `json:"height"`
				URL      string `json:"url"`
				HasAudio int    `json:"hasAudio"`
				Duration int    `json:"duration"`
				Vp8URL   string `json:"vp8Url"`
				H265URL  string `json:"h265Url"`
				Av1URL   string `json:"av1Url"`
			} `json:"image460sv"`
		} `json:"urls_resource,omitempty"`
		Raw struct {
			ID               string `json:"id"`
			URL              string `json:"url"`
			Title            string `json:"title"`
			Description      string `json:"description"`
			Type             string `json:"type"`
			Nsfw             int    `json:"nsfw"`
			UpVoteCount      int    `json:"upVoteCount"`
			DownVoteCount    int    `json:"downVoteCount"`
			CreationTs       int    `json:"creationTs"`
			Promoted         int    `json:"promoted"`
			IsVoteMasked     int    `json:"isVoteMasked"`
			HasLongPostCover int    `json:"hasLongPostCover"`
			Images           struct {
				Image700 struct {
					Width  int    `json:"width"`
					Height int    `json:"height"`
					URL    string `json:"url"`
				} `json:"image700"`
				Image460 struct {
					Width   int    `json:"width"`
					Height  int    `json:"height"`
					URL     string `json:"url"`
					WebpURL string `json:"webpUrl"`
				} `json:"image460"`
				ImageFbThumbnail struct {
					Width  int    `json:"width"`
					Height int    `json:"height"`
					URL    string `json:"url"`
				} `json:"imageFbThumbnail"`
				Image460Sv struct {
					Width    int    `json:"width"`
					Height   int    `json:"height"`
					URL      string `json:"url"`
					HasAudio int    `json:"hasAudio"`
					Duration int    `json:"duration"`
					Vp8URL   string `json:"vp8Url"`
					H265URL  string `json:"h265Url"`
					Av1URL   string `json:"av1Url"`
				} `json:"image460sv"`
			} `json:"images"`
			SourceDomain  string `json:"sourceDomain"`
			SourceURL     string `json:"sourceUrl"`
			CommentsCount int    `json:"commentsCount"`
			PostSection   struct {
				Name     string `json:"name"`
				URL      string `json:"url"`
				ImageURL string `json:"imageUrl"`
				WebpURL  string `json:"webpUrl"`
			} `json:"postSection"`
			Tags []interface{} `json:"tags"`
		} `json:"raw,omitempty"`
	} `json:"items"`
	ItemsDropped []interface{} `json:"items_dropped"`
	Stats        struct {
		DownloaderRequestBytes           int    `json:"downloader/request_bytes"`
		DownloaderRequestCount           int    `json:"downloader/request_count"`
		DownloaderRequestMethodCountGET  int    `json:"downloader/request_method_count/GET"`
		DownloaderResponseBytes          int    `json:"downloader/response_bytes"`
		DownloaderResponseCount          int    `json:"downloader/response_count"`
		DownloaderResponseStatusCount200 int    `json:"downloader/response_status_count/200"`
		DupefilterFiltered               int    `json:"dupefilter/filtered"`
		FinishReason                     string `json:"finish_reason"`
		FinishTime                       string `json:"finish_time"`
		ItemScrapedCount                 int    `json:"item_scraped_count"`
		LogCountDEBUG                    int    `json:"log_count/DEBUG"`
		LogCountINFO                     int    `json:"log_count/INFO"`
		MemusageMax                      int    `json:"memusage/max"`
		MemusageStartup                  int    `json:"memusage/startup"`
		RequestDepthMax                  int    `json:"request_depth_max"`
		ResponseReceivedCount            int    `json:"response_received_count"`
		SchedulerDequeued                int    `json:"scheduler/dequeued"`
		SchedulerDequeuedMemory          int    `json:"scheduler/dequeued/memory"`
		SchedulerEnqueued                int    `json:"scheduler/enqueued"`
		SchedulerEnqueuedMemory          int    `json:"scheduler/enqueued/memory"`
		StartTime                        string `json:"start_time"`
	} `json:"stats"`
	SpiderName string `json:"spider_name"`
}
type middleware func(next http.HandlerFunc) http.HandlerFunc

var tmpl = template.Must(template.ParseFiles("views/index.html"))
var startUrl = "https://agile-river-75797.herokuapp.com/crawl.json?spider_name=ninegag&start_requests=true"
var next *string

func (s *nineGagEntity) FormatPublishDate() string {
	date := time.Unix(int64(s.Items[0].CreationTs), 0)
	return date.Format(time.UnixDate)
}

func nextPage() string {
	logger.Debug(*next)
	return *next
}

func feedHandler(w http.ResponseWriter, r *http.Request) {
	s := &nineGagEntity{}

	resp, err := http.Get(nextPage())
	if err != nil {
		logger.Error(err)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		logger.Error(err)
		return
	}

	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		logger.Error(err)
		return
	}

	suffix := "&url=https://9gag.com/v1/group-posts/group/default/type/hot?after="
	nextCursor := fmt.Sprintf("%v%%2C%v%%2C%v&c=%d", s.Items[len(s.Items)-2].ID, s.Items[len(s.Items)-3].ID, s.Items[len(s.Items)-4].ID, s.Stats.ItemScrapedCount)
	pn := startUrl + suffix + nextCursor
	next = &pn

	err = tmpl.Execute(w, s)
	if err != nil {
		logger.Error(err)
		return
	}
}

func (s *nineGagEntity) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	resp, err := http.Get(startUrl)
	if err != nil {
		logger.Error(err)
		return
	}

	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&s)
	if err != nil {
		logger.Error(err)
		return
	}

	suffix := "&url=https://9gag.com/v1/group-posts/group/default/type/hot?after="
	nextCursor := fmt.Sprintf("%v%%2C%v%%2C%v&c=%d", s.Items[len(s.Items)-2].ID, s.Items[len(s.Items)-3].ID, s.Items[len(s.Items)-4].ID, s.Stats.ItemScrapedCount)
	pn := startUrl + suffix + nextCursor
	next = &pn

	err = tmpl.Execute(w, s)
	if err != nil {
		logger.Error(err)
		return
	}
}

func withLogging(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Debug("Connection from ", request.RemoteAddr)
		next.ServeHTTP(writer, request)
	}
}

func withTracing(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		logger.Debug("Tracing request for ", request.RequestURI)
		next.ServeHTTP(writer, request)
	}
}

func chainMiddleware(mw ...middleware) middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(writer http.ResponseWriter, request *http.Request) {
			last := final
			for i := len(mw) - 1; i > 0; i-- {
				last = mw[i](last)
			}
			last(writer, request)
		}
	}
}

func init() {
	// Init logger
	logger.SetFormatter(&logger.TextFormatter{ForceColors: true})
	logger.SetOutput(colorable.NewColorableStdout())
	logger.SetReportCaller(true)
	logger.SetLevel(logger.DebugLevel)
}

func main() {
	logger.Println("entry point...")
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*60, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	// Create Router
	router := mux.NewRouter()
	// Init chain middleware
	lt := chainMiddleware(withLogging, withTracing)
	// Structure pointer for use ServeHTTP method
	indexHandler := &nineGagEntity{}
	logger.Debug("Path to static: ", http.Dir("assets"))

	// Registered URLs
	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("assets"))))
	router.HandleFunc("/", lt(indexHandler.ServeHTTP))
	router.HandleFunc("/feed/", lt(feedHandler))

	// Set timeout to avoid Slowloris attacks( see https://en.wikipedia.org/wiki/Slowloris_(computer_security))
	server := &http.Server{
		Addr:         "127.0.0.1:27010",
		WriteTimeout: time.Second * 60,
		ReadTimeout:  time.Second * 60,
		IdleTimeout:  time.Second * 60,
		Handler:      router,
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := server.ListenAndServe(); err != nil {
			logger.Error(err)
		}
	}()
	c := make(chan os.Signal, 1)

	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal
	<-c

	// Create a deadline to wait for
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	defer cancel()

	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	server.Shutdown(ctx)
	logger.Warning("Shutting down...")
	os.Exit(0)
}
