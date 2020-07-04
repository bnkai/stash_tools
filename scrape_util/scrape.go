package main

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"sort"
	"time"

	"github.com/shurcooL/graphql"
	"gopkg.in/yaml.v2"
)

const defaultStashURL = "http://localhost:9998"

var verbose = false
var saveImages = false

type stringMap map[string]string

type ScrapedPerformer struct {
	Name         *string `graphql:"name" json:"name" yaml:"Name,omitempty"`
	Gender       *string `graphql:"gender" json:"gender" yaml:"Gender,omitempty"`
	URL          *string `graphql:"url" json:"url" yaml:"URL,omitempty"`
	Twitter      *string `graphql:"twitter" json:"twitter" yaml:"Twitter,omitempty"`
	Instagram    *string `graphql:"instagram" json:"instagram" yaml:"Instagram,omitempty"`
	Birthdate    *string `graphql:"birthdate" json:"birthdate" yaml:"Birthdate,omitempty"`
	Ethnicity    *string `graphql:"ethnicity" json:"ethnicity" yaml:"Ethnicity,omitempty"`
	Country      *string `graphql:"country" json:"country" yaml:"Country,omitempty"`
	EyeColor     *string `graphql:"eye_color" json:"eye_color" yaml:"EyeColor,omitempty"`
	Height       *string `graphql:"height" json:"height" yaml:"Height,omitempty"`
	Measurements *string `graphql:"measurements" json:"measurements" yaml:"Measurements,omitempty"`
	FakeTits     *string `graphql:"fake_tits" json:"fake_tits" yaml:"FakeTits,omitempty"`
	CareerLength *string `graphql:"career_length" json:"career_length" yaml:"CareerLength,omitempty"`
	Tattoos      *string `graphql:"tattoos" json:"tattoos" yaml:"Tatoos,omitempty"`
	Piercings    *string `graphql:"piercings" json:"piercings" yaml:"Piercings,omitempty"`
	Aliases      *string `graphql:"aliases" json:"aliases" yaml:"Aliases,omitempty"`
	Image        *string `graphql:"image" json:"image" yaml:"Image,omitempty"`
}

type ScrapedScene struct {
	Title      *string                  `graphql:"title" json:"title" yaml:"Title,omitempty"`
	Details    *string                  `graphql:"details" json:"details" yaml:"Details,omitempty"`
	URL        *string                  `graphql:"url" json:"url" yaml:"URL,omitempty"`
	Date       *string                  `graphql:"date" json:"date" yaml:"Date,omitempty"`
	Image      *string                  `graphql:"image" json:"image" yaml:"Image,omitempty"`
	Studio     *ScrapedSceneStudio      `graphql:"studio" json:"studio" yaml:"Studio,omitempty"`
	Movies     []*ScrapedSceneMovie     `graphql:"movies" json:"movies" yaml:"Movies,omitempty"`
	Tags       []*ScrapedSceneTag       `graphql:"tags" json:"tags" yaml:"Tags,omitempty"`
	Performers []*ScrapedScenePerformer `graphql:"performers" json:"performers" yaml:"Performers,omitempty"`
}

// uncomment fields when supported. should work with the toMaps function
type ScrapedScenePerformer struct {
	Name string `graphql:"name" json:"name" yaml:"Name,omitempty"`
	//	Gender       *string `graphql:"gender" json:"gender"`
	//	URL *string `graphql:"url" json:"url"`
	//	Twitter      *string `graphql:"twitter" json:"twitter"`
	//	Instagram    *string `graphql:"instagram" json:"instagram"`
	//	Birthdate    *string `graphql:"birthdate" json:"birthdate"`
	//	Ethnicity    *string `graphql:"ethnicity" json:"ethnicity"`
	//	Country      *string `graphql:"country" json:"country"`
	//	EyeColor     *string `graphql:"eye_color" json:"eye_color"`
	//	Height       *string `graphql:"height" json:"height"`
	//	Measurements *string `graphql:"measurements" json:"measurements"`
	//	FakeTits     *string `graphql:"fake_tits" json:"fake_tits"`
	//	CareerLength *string `graphql:"career_length" json:"career_length"`
	//	Tattoos      *string `graphql:"tattoos" json:"tattoos"`
	//	Piercings    *string `graphql:"piercings" json:"piercings"`
	//	Aliases      *string `graphql:"aliases" json:"aliases"`
}

type ScrapedSceneStudio struct {
	Name string `graphql:"name" json:"name" yaml:"Name,omitempty"`
	//URL  *string `graphql:"url" json:"url"`
}

type ScrapedSceneMovie struct {
	Name string `graphql:"name" json:"name" yaml:"Name,omitempty"`
	//	Aliases  string  `graphql:"aliases" json:"aliases"`
	//	Duration string  `graphql:"duration" json:"duration"`
	//	Date     string  `graphql:"date" json:"date"`
	//	Rating   string  `graphql:"rating" json:"rating"`
	//	Director string  `graphql:"director" json:"director"`
	//	Synopsis string  `graphql:"synopsis" json:"synopsis"`
	//URL *string `graphql:"url" json:"url"`
}

type ScrapedSceneTag struct {
	Name string `graphql:"name" json:"name" yaml:"Name,omitempty"`
}

type ScrapePerformerData struct {
	PerformerURL  string           `yaml:"PerformerURL"`
	PerformerData ScrapedPerformer `yaml:"PerformerData,omitempty"`
}
type ScrapeSceneData struct {
	SceneURL  string       `yaml:"SceneURL"`
	SceneData ScrapedScene `yaml:"SceneData,omitempty"`
}
type ScrapeData struct {
	PData []ScrapePerformerData `yaml:"Performers,omitempty"`
	SData []ScrapeSceneData     `yaml:"Scenes,omitempty"`
}

// convert ScapedPerformer struct to a string Map
func (p *ScrapedPerformer) toMap() *stringMap {
	performerMap := make(stringMap)
	v := reflect.ValueOf(*p)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).Interface().(*string) != nil {
			performerMap[t.Field(i).Name] = *(v.Field(i).Interface().(*string))
		}
	}

	return &performerMap
}

// print a string Map
// for the image field, if its a valid base64 file its md5 is printed instead
func (m *stringMap) printMap() {
	for k, v := range *m {
		if k == "Image" {
			md5, _, err := MD5FromBase64(v)
			if err != nil {
				fmt.Printf("Error decoding base64 :%s\n", err)
				v = ""
			} else {
				v = md5
			}

		}
		fmt.Printf(" %s = %s\n", k, v)
	}
}

func equalStringMaps(toVerify, toScrape *stringMap, verbose bool) bool {
	if toVerify == nil || toScrape == nil {
		if verbose {
			fmt.Printf("Maps must be != nil\n")
		}
		return false
	}
	equal := true
	matches := 0

	if len(*toVerify) != len(*toScrape) {
		if verbose {
			fmt.Printf("Length of maps is different  wanted %d != %d tested", len(*toVerify), len(*toScrape))
		}
		return false
	}
	for k, v := range *toVerify {
		if v != (*toScrape)[k] {
			equal = false
			if !verbose {
				return false
			}
			fmt.Printf("Values for key %s differ.\nWanted: %s\nFound: %s", k, v, (*toScrape)[k])
		} else {
			matches++
		}
	}
	if verbose {
		fmt.Printf(" %d / %d keys match\n", matches, len(*toVerify))
	}

	return equal
}

func sortStringMapSliceByName(mapSlice *[]stringMap) {
	sort.SliceStable(*mapSlice, func(i, j int) bool { return (*mapSlice)[i]["Name"] < (*mapSlice)[j]["Name"] })
}

func equalStringMapSlices(slice, scrapedSlice *[]stringMap) bool {
	if len(*slice) != len(*scrapedSlice) {
		return false
	} else {
		sortStringMapSliceByName(slice)
		sortStringMapSliceByName(scrapedSlice)

		for i := range *slice {
			if !equalStringMaps(&(*slice)[i], &(*scrapedSlice)[i], verbose) {
				return false

			}

		}
	}
	return true
}

func (s *ScrapedScene) toMaps(sceneSimple, studio stringMap, movies, tags, performers *[]stringMap) {
	v := reflect.ValueOf(*s)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		switch t.Field(i).Name {
		case "Studio":
			reflected := v.Field(i).Interface().(*ScrapedSceneStudio)
			if reflected == nil {
				continue
			}
			stV := reflect.ValueOf(*reflected)
			stT := stV.Type()
			for j := 0; j < stV.NumField(); j++ {
				switch stT.Field(j).Name {
				case "Name":
					name := stV.Field(j).Interface().(string)
					if name != "" {
						studio["Name"] = name
					}
				default:
					if stV.Field(j).Interface().(*string) != nil {
						value := *(stV.Field(j).Interface().(*string))
						studio[stT.Field(j).Name] = value
					}
				}
			}

		case "Movies":
			for _, reflected := range v.Field(i).Interface().([]*ScrapedSceneMovie) {
				mV := reflect.ValueOf(*reflected)
				mT := mV.Type()
				movie := make(stringMap)
				for j := 0; j < mV.NumField(); j++ {
					switch mT.Field(j).Name {
					case "URL":
						value := mV.Field(j).Interface().(*string)
						if value != nil {
							movie["URL"] = *value
						}
					default: // default, treat each field as string
						if mV.Field(j).Interface().(string) != "" {
							movie[mT.Field(j).Name] = (mV.Field(j).Interface().(string))
						}
					}
				}

				*movies = append(*movies, movie)
			}
		case "Tags":
			for _, reflected := range v.Field(i).Interface().([]*ScrapedSceneTag) {
				tV := reflect.ValueOf(*reflected)
				tT := tV.Type()
				tag := make(stringMap)
				for j := 0; j < tV.NumField(); j++ {
					switch tT.Field(j).Name {
					case "Name":
						name := tV.Field(j).Interface().(string)
						if name != "" {
							tag["Name"] = name
						}
					default: // default, treat each field as *string
						if tV.Field(j).Interface().(*string) != nil {
							tag[tT.Field(j).Name] = *(tV.Field(j).Interface().(*string))
						}
					}
				}

				*tags = append(*tags, tag)
			}

		case "Performers":
			for _, reflected := range v.Field(i).Interface().([]*ScrapedScenePerformer) {
				pV := reflect.ValueOf(*reflected)
				pT := pV.Type()
				performer := make(stringMap)
				for j := 0; j < pV.NumField(); j++ {
					switch pT.Field(j).Name {
					case "Name":
						name := pV.Field(j).Interface().(string)
						if name != "" {
							performer["Name"] = name
						}
					default: // default , treat each field as *string
						if pV.Field(j).Interface().(*string) != nil {
							performer[pT.Field(j).Name] = *(pV.Field(j).Interface().(*string))
						}
					}
				}

				*performers = append(*performers, performer)
			}
		default: // default treat each field as a *string
			if v.Field(i).Interface().(*string) != nil {
				value := *(v.Field(i).Interface().(*string))
				sceneSimple[t.Field(i).Name] = value
			}
		}
	}
}

func getStashClient(url string) *graphql.Client {
	return graphql.NewClient(url+"/graphql", nil)
}

func reloadStashScrapers(stash string) bool {
	client := getStashClient(stash)
	var m struct {
		ReloadScrapers graphql.Boolean `graphql:"reloadScrapers"`
	}

	err := client.Mutate(context.Background(), &m, nil)
	if err != nil {
		fmt.Printf("Error reloading scrapers:%s\n", err)
	}
	return bool(m.ReloadScrapers)
}

func scrapePerformerByUrl(url string, stash string, saveImg bool) *ScrapedPerformer {
	client := getStashClient(stash)
	var q struct {
		ScrapedPerformer `graphql:"scrapePerformerURL(url: $performerUrl)" yaml:"Performer"`
	}

	vars := map[string]interface{}{
		"performerUrl": graphql.String(url),
	}

	err := client.Query(context.Background(), &q, vars)
	if err != nil {
		fmt.Printf("Error quering stash :%s", err)
		return nil
	}

	// instead of getting the whole image content
	// use the md5 of the image
	if q.Image != nil {
		md5, data, err := MD5FromBase64(*q.Image)
		if err == nil {
			*q.Image = md5
			if saveImg {
				err = ioutil.WriteFile(md5+getExt(&data), data, 0644)
				if err != nil {
					fmt.Printf("Error writing image file :%s \n", err)
				}
			}

		}
	}
	return &q.ScrapedPerformer

}

func scrapePerformersByUrl(urls []string, stash string, saveImg bool) []byte {
	var performers []ScrapePerformerData

	for _, url := range urls {
		performer := scrapePerformerByUrl(url, stash, saveImages)
		t := ScrapePerformerData{PerformerURL: url, PerformerData: *performer}
		performers = append(performers, t)
	}
	tt := ScrapeData{PData: performers}
	dataPerformer, _ := yaml.Marshal(tt)
	return dataPerformer
}

func scrapeSceneByUrl(url string, stash string, saveImg bool) *ScrapedScene {
	client := getStashClient(stash)

	var q struct {
		ScrapedScene `graphql:"scrapeSceneURL(url: $sceneUrl)" yaml:"Scene"`
	}

	vars := map[string]interface{}{
		"sceneUrl": graphql.String(url),
	}

	err := client.Query(context.Background(), &q, vars)
	if err != nil {
		fmt.Printf("Error quering stash :%s", err)
		return nil
	}

	// instead of getting the whole image content
	// use the md5 of the image
	if q.Image != nil {
		md5, data, err := MD5FromBase64(*q.Image)
		if err == nil {
			*q.Image = md5
		}
		if saveImg {
			err = ioutil.WriteFile(md5+getExt(&data), data, 0644)
			if err != nil {
				fmt.Printf("Error writing image file :%s \n", err)
			}
		}

	}
	return &q.ScrapedScene
}

func scrapeScenesByUrl(urls []string, stash string, saveImg bool) []byte {
	var scenes []ScrapeSceneData
	for _, url := range urls {
		scene := scrapeSceneByUrl(url, stash, saveImg)
		t := ScrapeSceneData{SceneURL: url, SceneData: *scene}
		scenes = append(scenes, t)

	}

	tt := ScrapeData{SData: scenes}

	dataScene, _ := yaml.Marshal(tt)
	return dataScene

}

func MD5FromBase64(imageString string) (string, []byte, error) {
	if imageString == "" {
		return "", nil, fmt.Errorf("empty image string")
	}

	regex := regexp.MustCompile(`^data:.+\/(.+);base64,(.*)$`)
	matches := regex.FindStringSubmatch(imageString)
	var encodedString string
	if len(matches) > 2 {
		encodedString = regex.FindStringSubmatch(imageString)[2]
	} else {
		encodedString = imageString
	}

	data, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		return "", nil, err
	}
	return fmt.Sprintf("%x", md5.Sum(data)), data, nil

}
func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func getExt(data *[]byte) string {
	var extension = ""
	if data != nil {
		mime := http.DetectContentType(*data)
		switch mime {
		case "image/jpeg":
			extension = ".jpg"
		case "image/bmp":
			extension = ".bmp"
		case "image/png":
			extension = ".png"
		case "image/gif":
			extension = ".gif"
		case "image/webp":
			extension = ".webp"
		default:
			extension = ""
		}

	}
	return extension
}

func generate(scrapeDefault bool, urls []string, output *string, stash *string) {
	stashURL := defaultStashURL
	if stash != nil {
		stashURL = *stash
	}

	mode := "scenes"
	if !scrapeDefault {
		mode = "performers"
	}

	outputFile := mode + "_data_" + time.Now().Format("20060102_150405") + ".yml"
	if output != nil && *output != "" {
		outputFile = *output
	}

	// reload scrapers to make sure they are loaded
	reloadStashScrapers(*stash)

	var data []byte
	if !scrapeDefault {
		data = scrapePerformersByUrl(urls, stashURL, saveImages)
	} else {
		data = scrapeScenesByUrl(urls, stashURL, saveImages)
	}
	fmt.Printf("Writing to file %s %d bytes\n", outputFile, len(data))
	err := ioutil.WriteFile(outputFile, data, 0644)
	if err != nil {
		fmt.Printf("Error writing file :%s \n", err)
	}
}

var (
	action   = flag.String("action", "gen", "Mode of operation Generate (-action=\"gen\") or Test (-action=\"test\").")
	file     = flag.String("file", "", "Filename to store data to when in Generate mode  or read data from when in Test mode (-file=\"filename.yml\")")
	stash    = flag.String("stash", defaultStashURL, "Stash URL to use (-stash=\"http://mystash.com:9998\")")
	scrape   = flag.String("scrape", "scene", "Scrape performer if \"perf\", scene otherwise")
	verb     = flag.Bool("verb", false, "Verbose mode. Set to true (-verb=true) to enable")
	urlsFile = flag.String("urls", "", "Optional urls file to use for generating")
	images   = flag.Bool("img", false, "Save images. Set to true (-img=true or -img) to enable")
)

func main() {
	flag.Parse()
	urls := flag.Args()

	scrapeDefault := true
	scrapeFor := "scenes"
	if scrape != nil && (*scrape) == "perf" {
		scrapeFor = "performers"
		scrapeDefault = false
	}

	verbose = *verb
	saveImages = *images

	if *action == "gen" { // generate yaml files for the urls given
		if len(urls) < 1 && (urlsFile == nil || *urlsFile == "") {
			fmt.Println("No URLs where given. Use like this: ./scrape -action=\"gen\" \"http://url.to.scrape/1\" \"http://url.to.scrape/2\" or like this: ./scrape -action=\"gen\" -urls=\"scene_urls.urls\"")
			return
		}
		if !(urlsFile == nil || *urlsFile == "") {
			urlsFromFile, err := readLines(*urlsFile)
			if err == nil {
				urls = append(urls, urlsFromFile...)
			}
		}
		fmt.Printf("\nUsing (%d) url/s to scrape for %s\n", len(urls), scrapeFor)
		generate(scrapeDefault, urls, file, stash)
	} else { // compare the data from the file given to the newly scraped ones
		if file != nil && *file != "" {
			yamlFile, err := ioutil.ReadFile(*file)
			if err != nil {
				fmt.Printf("Error: %s opening file: %s\n", err, *file)
				return
			}

			data := ScrapeData{}
			err = yaml.Unmarshal(yamlFile, &data)
			if err != nil {
				fmt.Printf("Unmarshal: %v\n", err)
				return
			}

			// reload scrapers to make sure they are loaded
			reloadStashScrapers(*stash)

			//fetch and compare performer data first
			for _, performer := range data.PData {
				verifyPerformerMap := performer.PerformerData.toMap()
				scraped := scrapePerformerByUrl(performer.PerformerURL, *stash, false)
				if scraped == nil {
					fmt.Printf("No scraped data for %s\n ", performer.PerformerURL)
					continue
				}
				if equalStringMaps(verifyPerformerMap, scraped.toMap(), verbose) {
					fmt.Printf("Data for %s matches\n", performer.PerformerURL)
				} else {
					fmt.Printf("Data for %s differs\n", performer.PerformerURL)
				}

			}

			// now compare scene data
			for _, scene := range data.SData {
				equalScene := true

				simpleMap := make(stringMap)
				studioMap := make(stringMap)
				performersMap := make([]stringMap, 0, 0)
				tagsMap := make([]stringMap, 0, 0)
				moviesMap := make([]stringMap, 0, 0)

				scene.SceneData.toMaps(simpleMap, studioMap, &moviesMap, &tagsMap, &performersMap)

				scraped := scrapeSceneByUrl(scene.SceneURL, *stash, false)
				if scraped == nil {
					fmt.Printf("Data differs. No scraped data for %s\n ", scene.SceneURL)
					continue

				}
				//get maps from the  scraped data and compare
				scrapedSimpleMap := make(stringMap)
				scrapedStudioMap := make(stringMap)
				scrapedPerformersMap := make([]stringMap, 0, 0)
				scrapedTagsMap := make([]stringMap, 0, 0)
				scrapedMoviesMap := make([]stringMap, 0, 0)

				scraped.toMaps(scrapedSimpleMap, scrapedStudioMap, &scrapedMoviesMap, &scrapedTagsMap, &scrapedPerformersMap)

				// compare simple fields title,date,details .....
				if !equalStringMaps(&simpleMap, &scrapedSimpleMap, verbose) {
					equalScene = false
				}
				// compare studio
				if !equalStringMaps(&studioMap, &scrapedStudioMap, verbose) {
					equalScene = false
				}

				// compare performers

				if !equalStringMapSlices(&performersMap, &scrapedPerformersMap) {
					equalScene = false
				}

				// compare tags
				if !equalStringMapSlices(&tagsMap, &scrapedTagsMap) {
					equalScene = false

				}

				// compare movies

				if !equalStringMapSlices(&moviesMap, &scrapedMoviesMap) {
					equalScene = false
				}

				if equalScene {
					fmt.Printf("Scene data for %s matches\n", scene.SceneURL)
				} else {
					fmt.Printf("!!! Scene data for %s differs\n", scene.SceneURL)
				}

			}

		} else {
			fmt.Println("No input file given! : -file=\"filename.yml\"")
		}
	}
}
