
`./scrape -action="gen" -scrape="perf" -file="performers_freeones.yml" "https://www.freeones.xxx/asuna-fox/profile" "https://www.freeones.xxx/candice-dare/profile" "https://www.freeones.xxx/misty-quinn/profile" "https://www.freeones.xxx/veronica-rodriguez/profile"`

creates a performers_freeones.yml file with the performer info scraped from the given urls


`/scrape -scrape="scenes" -file="bangbros_scenes.yml" "https://bangbros.com/video3402241/stretching-out-that-tight-pussy" "https://bangbros.com/video3409739/a-splashy-public-fuck" "https://bangbros.com/video3409685/my-maid-loves-cock" "https://bangbros.com/video3409667/public-fun-with-alysia" "https://bangbros.com/video3409645/kiras-sexy-vacation-day-2" "https://bangbros.com/video3409673/step-sister-fuck-punishment"`

creates a bangbros_scenes.yml file with scene info scraped from the given urls

`/scrape -file="various_scenes.yml" "https://bangbros.com/video3402241/stretching-out-that-tight-pussy" "https://www.newsensations.com/tour_ns/updates/Natalie-Brooks-Step-Sister-4k-Black-Cock-Swallowing-Black-Brother-Thanksgiving-BBC.html" "https://www.lookathernow.com/scene/4408588/mukbang-her"  "https://www.mrpov.com/scene/mrpov/my-favorite-fan-ever"  "https://www.naughtyamerica.com/scene/charlotte-sins-fucks-friends-brother-25711"`

creates a various_scenes.yml file containing info from various scene scrapers.

`/scrape -file="various_scenes.yml" -urls="urls_to_scrape.urls"`

same as above but reading the urls from a file (one url per line)

`./scrape -action="test" -file="performers_freeones.yml"`

reads the performers_freeones.yml file and compares each entry there with "fresh" scraped data from stash



**TODO**
 authentication is not supported. for now you need to disable it from stash
 
**NOTES**
 some sites use caching for images (with different quality) so while images are visually the same the md5 can be different 
 
 use `-stash="http://mystash.com:9998"` to set stash's url , default is  `"http://localhost:9998"`
