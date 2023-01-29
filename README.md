# steamprom

steamprom is a go microservice that exposes your Steam game playtime via
Prometheus metrics.

This service runs at https://steamprom.apps.notk.ai/.

## API

`GET /id/{apiKey}/{steamID}`  
`apiKey` - your Steam Web API key  
`steamID` - your Steam ID in int64 format

This will return all of your known Steam game playtime in metrics format. It is
separated by platform (Windows, macOS, and Linux) as well as the Steam API can,
and it is assumed that all unaccounted for playtime happened on Windows (which
is statistically the most likely case).