// My hope is this code is so shit that I'm not allowed to write golang ever again
package main

import (
	twitterscraper "github.com/regalialong/twitter-scraper"
	"github.com/rs/zerolog/log"
	"net/http"
	"regexp"
	"strings"
)

// TODO Not being allowed to view the tweet due to NSFW is also a 404 for whatever ungodly reason fuck you musk, we should see if we can't pass a cookie or detect if the tweet is NSFW.
// The problem is that we don't have access to the object in the first place
func fetchSingleTweet(ID string, scraper *twitterscraper.Scraper) *twitterscraper.Tweet {
	tweet, err := scraper.GetTweet(ID)
	if err != nil {
		if strings.Contains(err.Error(), "404") || strings.Contains(err.Error(), "tweet with ID") {
			log.Debug().Str("Id", ID).Msg("Fetch failed because of 404 (deleted or NSFW)")
			return nil
		}
		log.Debug().Str("Id", ID).Msg("Fetch failed: " + err.Error())
		return nil
	}
	if tweet.IsQuoted == true && tweet.QuotedStatus == nil {
		tweet.QuotedStatus = fetchSingleTweet(tweet.QuotedStatusID, scraper)
	}
	if tweet.IsRetweet == true && tweet.RetweetedStatus == nil {
		tweet.RetweetedStatus = fetchSingleTweet(tweet.RetweetedStatusID, scraper)
	}
	if tweet.IsReply == true && tweet.InReplyToStatus == nil {
		tweet.InReplyToStatus = fetchSingleTweet(tweet.InReplyToStatusID, scraper)
	}

	return tweet
}
func removeShort(statusStr string, urlobj []twitterscraper.Urls) string {
	for _, link := range urlobj {
		statusStr = strings.Replace(statusStr, link.ShortURL, link.ExpandedURL, 1)
	}

	// Since we are not using the proper Twitter API, tweets have a t.co reference t o themselves at the end.
	// We can't simply cut off the end since t.co URLs vary in size,
	// so we have to sadly parse them to remove.
	statusStr = reRemoveShort(statusStr)
	return statusStr
}

func reRemoveShort(statusStr string) string {
	re := regexp.MustCompile("\\bhttps?://t\\.co/[a-zA-Z0-9]+(\\?.*)?\\b")
	tcos := re.FindAllString(statusStr, -1)
	for _, link := range tcos {
		res, err := http.Get(link)

		if err != nil {
			panic(err)
		}
		// Tweets sometimes contain references to themselves or related context (retweet, quote tweet etc.)
		// Since we handle that ourselves, we don't require them.
		if res.Request.URL.Host == "twitter.com" {
			statusStr = strings.Replace(statusStr, link, "", -1)
		}
	}
	return statusStr
}

// returns <a href="https://twitter.com/{handle}">@{handle}</a>
func handleToHyperlink(original string) (parsed string) {
	re := regexp.MustCompile("@([\\w]+)")
	handles := re.FindAllString(original, -1)
	parsed = original
	for i, handle := range handles {
		if i != 0 && handle == handles[i-1] { // This is a stupid hack, but I don't know an alternative.
			continue
		}
		hyperlinkHandle := "<a href=\"https://twitter.com/" + handle[1:] + "\">@" + handle[1:] + "</a>"
		parsed = strings.Replace(parsed, handle, " "+hyperlinkHandle, -1)
	}
	return parsed
}

func ProcessTweetContent(tweet *twitterscraper.TweetResult) string {
	statusStr := ""

	// Doing string concatenation while growing up with Python
	// is like missing your ex, I just want fstrings back...
	// Addendum: No, sprintf is not better imo.
	if tweet.IsReply {
		statusStr = "Reply to  @" + tweet.InReplyToStatus.Username + ":<br>" + tweet.InReplyToStatus.Text + "<br><br>↳<br> "

	}

	if tweet.IsRetweet {
		replacementCase := "RT @" + tweet.RetweetedStatus.Username + ":"
		statusStr = strings.Replace(tweet.Text, replacementCase, "@"+tweet.Username+" 🔁 @"+tweet.RetweetedStatus.Username+"<br>", -1)
	} else {
		statusStr = statusStr + "@" + tweet.Username + ": " + tweet.Text + "<br>"
	}

	statusStr = handleToHyperlink(statusStr)
	statusStr = removeShort(statusStr, tweet.URLs)

	statusStr = statusStr + " <br><a href=\"" + tweet.PermanentURL + "\">Source</a>"
	return statusStr

}
