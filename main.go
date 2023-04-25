package main

import (
	"context"
	"encoding/json"
	twitterscraper "github.com/regalialong/twitter-scraper"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	wg            sync.WaitGroup
	configuration                         = configf{}
	barrier                               = make(chan struct{}, len(configuration.Handles))
	scraper       *twitterscraper.Scraper = nil
)

type configf struct {
	BASEURL  string   `json:"BASEURL"`
	PASSWORD string   `json:"PASSWORD"`
	REQDELAY int64    `json:"REQDELAY"`
	PROXYURL string   `json:"PROXYURL,omitempty"`
	USERNAME string   `json:"USERNAME"`
	Handles  []string `json:"handles"`
}

func grabUrlToLocal(img string) string {
	fileURL, err := url.Parse(img)
	if err != nil {
		panic(err)
	}
	path := fileURL.Path
	segments := strings.Split(path, "/")
	fileName := segments[len(segments)-1]

	out, err := os.Create("files/" + fileName)
	if err != nil {
		panic(err)
	}

	resp, err := http.Get(img)
	if err != nil {
		panic(err)
	}

	_, err = io.Copy(out, resp.Body)
	return fileName
}

func loadConfig() (configuration configf) {
	content, err := os.ReadFile("config.json")
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &configuration)
	if err != nil {
		panic(err)
	}

	return configuration
}

func init() {
	configuration = loadConfig()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(zerolog.TraceLevel)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	// TODO Proxy should be configured on twitter-scraper level
	if configuration.PROXYURL != "" {
		proxyUrl, _ := url.Parse(configuration.BASEURL)
		log.Debug().Msg("Loading proxy URL: " + proxyUrl.String())
		http.DefaultTransport = &http.Transport{
			Proxy: http.ProxyURL(proxyUrl),
		}
	}
	dups := removeDuplicate(configuration.Handles)
	if len(dups) != len(configuration.Handles) {
		log.Panic().Msg("Duplicates in handles config!")
	}
}

func main() {
	handlecount := len(configuration.Handles)
	if handlecount < 1 {
		log.Fatal().Msg("No handles")
		panic("You need to add at least once handle to poll!")

	}

	scraper = twitterscraper.New()
	wg.Add(handlecount)
	log.Info().Msg("Spinning " + strconv.Itoa(handlecount) + " threads")
	for i, user := range configuration.Handles {
		log.Info().Msg("Loading @" + user)
		go mainlogicloop(user, i)
	}

	for i := 0; i < len(configuration.Handles); i++ {
		<-barrier
	}
	log.Trace().Msg("We passed barrier")

	wg.Done()

	// Block main
	wg.Wait()

}

func mainlogicloop(handle string, id int) {
	// TODO logic here should probably be moved to twscrp.go instead

	var origTweet *twitterscraper.TweetResult
	twtCount := -1

	for {
		twtCount++
		tweets := scraper.GetTweets(context.Background(), handle, 1)
		for tweet := range tweets {
			if tweet.Error != nil {
				err := tweet.Error.Error()
				log.Error().Str("Handle", handle).Msg(err)
				if strings.Contains(err, "429") || strings.Contains(err, "401") {
					log.Trace().Int("Count", twtCount).Str("Handle", handle).Msg("Ratelimited, rotating session...")
					scraper = twitterscraper.New()
					continue
				}
			}

			if origTweet == nil {
				log.Debug().Str("Handle", handle).Int("Thread number", id).Msg("Setting initial tweet.")
				origTweet = tweet

				// Signal we have initialized and wait for others.
				barrier <- struct{}{}
				continue
			}
			if tweet.ID == origTweet.ID {
				log.Debug().Int("Count", twtCount).Str("Handle", handle).Msg("Tweet is unchanged...")
				time.Sleep(time.Duration(configuration.REQDELAY) * time.Second)
				continue
			}

			var media []string
			if tweet.Photos != nil {
				for _, img := range tweet.Photos {
					media = append(media, grabUrlToLocal(img))
				}
			}

			if tweet.Videos != nil {
				for _, video := range tweet.Videos {
					media = append(media, grabUrlToLocal(video.URL))
				}
			}

			if tweet.IsReply {
				if tweet.InReplyToStatus == nil {
					log.Debug().Str("Handle", handle).Str("ReplyID", tweet.InReplyToStatusID).Msg("Tweet is a reply but lacks reply context, fetching context")
					tweet.InReplyToStatus = fetchSingleTweet(tweet.InReplyToStatusID, scraper)
					if tweet.InReplyToStatus != nil {
						log.Debug().Str("Handle", handle).Str("ReplyID", tweet.InReplyToStatusID).Str("Original ID", tweet.ID).Msg("Successfully recovered reply context")
					}
				}
				if tweet.InReplyToStatus != nil {
					if tweet.InReplyToStatus.Photos != nil {
						for _, img := range tweet.InReplyToStatus.Photos {
							media = append(media, grabUrlToLocal(img))
						}

						if tweet.InReplyToStatus.Videos != nil {
							for _, video := range tweet.Videos {
								media = append(media, grabUrlToLocal(video.URL))
							}
						}
					}
				} else {
					log.Warn().Str("Handle", handle).Str("ReplyID", tweet.InReplyToStatusID).Str("Original ID", tweet.ID).Msg("Reply context does not exist despite mitigations, setting as non-reply")
					tweet.IsReply = false
				}
			}

			if tweet.IsQuoted {
				if tweet.QuotedStatus == nil {
					log.Debug().Str("Handle", handle).Msg("Tweet is a quote but lacks quote context, fetching context")
					tweet.QuotedStatus = fetchSingleTweet(tweet.QuotedStatusID, scraper)
					if tweet.QuotedStatus != nil {
						log.Debug().Str("Handle", handle).Str("Quote ID", tweet.QuotedStatusID).Str("Original ID", tweet.ID).Msg("Successfully recovered quote context")
					}
				}
				if tweet.QuotedStatus != nil {
					if tweet.QuotedStatus.Photos != nil {
						for _, img := range tweet.QuotedStatus.Photos {
							media = append(media, grabUrlToLocal(img))
						}

						if tweet.QuotedStatus.Videos != nil {
							for _, video := range tweet.Videos {
								media = append(media, grabUrlToLocal(video.URL))
							}
						}
					}
				} else {
					log.Warn().Str("Handle", handle).Str("Quote ID", tweet.QuotedStatusID).Str("Original ID", tweet.ID).Msg("Quote context does not exist despite mitigations, setting as non-Quote")
					tweet.IsQuoted = false
				}
			}

			statusStr := ProcessTweetContent(tweet)

			var ids []string

			for _, file := range media {
				ids = append(ids, send_file("files/"+file).Id)
			}

			body := StatusParameters{
				Status:      statusStr,
				ContentType: "text/html",
				MediaIds:    ids,
				Visibility:  "unlisted"}
			status(body)
			log.Debug().Str("Handle", handle).Msg("Tweeted: " + body.Status)
			origTweet = tweet

		}
	}
}
