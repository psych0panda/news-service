package main

import (
	"encoding/json"
	"fmt"
	"github.com/mattn/go-colorable"
	logger "github.com/sirupsen/logrus"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"
)

type NineGagEntity struct {
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

var tmpl = template.Must(template.ParseFiles("views/index.html"))
var startUrl = "https://agile-river-75797.herokuapp.com/crawl.json?spider_name=ninegag&start_requests=true"
var capacityPage = 0

func (s *NineGagEntity) FormatPublishDate() string {
	date := time.Unix(int64(s.Items[0].CreationTs), 0)
	return date.Format(time.UnixDate)
}

func (s *NineGagEntity) BuildCursor() string {
	cursor := fmt.Sprintf("%v%%2C%v%%2C%v&c=%d", s.Items[len(s.Items)-1].ID, s.Items[len(s.Items)-2].ID, s.Items[len(s.Items)-3].ID, capacityPage+40)
	capacityPage += 40
	return cursor
}

func (s *NineGagEntity) NextPage() string {
	suffix := fmt.Sprintf("&url=https://9gag.com/v1/group-posts/group/default/type/hot?after=%s", s.BuildCursor())
	nextPage := startUrl + suffix
	return nextPage
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	nineGagStruct := &NineGagEntity{}

	u, err := url.Parse(r.URL.String())
	if err != nil {
		logger.Error(err)
		return
	}

	params := u.Query()
	page := params.Get("page")
	if page != "" {
		startUrl = nineGagStruct.NextPage()
	} else {
		page = "1"
	}

	//resp, err := http.Get(startUrl)
	file, err := ioutil.ReadFile("response.json")
	if err != nil {
		logger.Error(err)
		return
	}
	_ = json.Unmarshal([]byte(file), nineGagStruct)

	//defer resp.Body.Close()
	//if resp.StatusCode != 200 {
	//	logger.Error("Exception in response: ", http.StatusInternalServerError)
	//	return
	//}

	//err = json.NewDecoder(resp.Body).Decode(&nineGagStruct.Items)
	//if err != nil {
	//	logger.Error("Exception in json decode: ", err)
	//	return
	//}

	err = tmpl.Execute(w, nineGagStruct)
	if err != nil {
		logger.Error(err)
		return
	}
}

func init() {
	logger.SetFormatter(&logger.TextFormatter{ForceColors: true})
	logger.SetOutput(colorable.NewColorableStdout())
	logger.SetReportCaller(true)
	logger.SetLevel(logger.DebugLevel)
}

func main() {
	logger.Println("entry point...")
	port := os.Getenv("PORT")
	if port == "" {
		port = "27010"
	}
	mux := http.NewServeMux()

	fs := http.FileServer(http.Dir("assets"))

	mux.Handle("/assets/", http.StripPrefix("/assets/", fs))
	mux.HandleFunc("/", indexHandler)

	err := http.ListenAndServe(":"+port, mux)
	if err != nil {
		logger.Fatal(err)
	}
}
