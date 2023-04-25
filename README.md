# Furfetcher 

This project scrapes the Twitter Frontend API using a modified version of [n0madic/twitter-scraper](https://github.com/n0madic/twitter-scraper) 
and crossposts it to a Pleroma/Akkoma Mastodon API compatible instance. Intended for furry art.

While the code here should work, this project has failed, so I'm putting it here because I give up.

`config.json` should be like this:
```json
{
"BASEURL": "https://yourinstancehere.org",
"USERNAME": "botaccountusername",
"PASSWORD": "botaccountpassword",
"REQDELAY": 10,
"handles": [
"twitter_handle_from_user",
"sifyro"
]
}
```
~~Please don't run this though for reasons at the bottom~~
~~Sorry by the way for the comments in this place~~

## Autopsy
### Twitter
#### Scaling
This projects doesn't scale, meaning you can't put a lot of handles before you get ratelimited.
This is the biggest reason why I am dropping it. I don't understand the ratelimit of the Twitter API which makes it impossible to scale this project up to more accounts.

While I might be able to use accounts and proxies to bypass this, that becomes very costly for what is essentially a passion project.

#### NSFW Changes
While right now I'm not interested in NSFW content, the Frontend API cannot access NSFW content without authentication. This has happened while I was in the middle of developing this. We can't probe this and we miss context to why this happens since the API effectively handles it as the Tweet not existing.

### Fediverse and Crossposting
I haven't asked Admins opinions on this but from reading up, Fedi doesn't enjoy crossposting from Twitter especally since an account would effectively be a zombie doing nothing. This is why this project cannot be federated, therefore limiting the usability of it for other people. 

That makes it difficult to involve help (either by other people offering themselves as proxy or upstream) from others.

### Codebase
This code is really messy, it requires reworks in a lot of places. A lot of logic intended for twscrp is in main. I don't enjoy looking at it or working with it.
#### Lack of Testing
I would also have liked to add testing to ensure some text parsing functions don't break but never got around to it.
#### No OAuth
I didn't know how to do this per Scopes, so this code simply uses the username and password instead. That's not good.

### Soft-dependency on modified library
This uses the modified library of mine for convenience, but I probably won't maintain that much. You can probably rework this back to use the proper library, but I don't want to maintain this anymore.