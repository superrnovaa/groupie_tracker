package models

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"html/template"
	"os"
	"strconv"
	"strings"
	
)

var artistTemp []BandInfo
var locTemp = map[string][]BandInfo{}

type ApiData struct {
	AllBands     []BandInfo
	DisplayBands []BandInfo
	filterBands  []BandInfo
}

func (self *ApiData) FeedApi() {
	/*
		Method of ApiData
		Extract the data from the heroku api to self.AllBands and self.DisplayBands.
		Current content of self.AllBands and self.DisplayBands will be erased and replaced by new data.
	*/
	//artistTemp := []BandInfo{}
	//locTemp := map[string][]BandInfo{}
	log.Printf("[INFO] - Extracting data from API...\n")
	for _, api := range []string{"artists", "relation"} {
		req, err := http.Get("https://groupietrackers.herokuapp.com/api/" + api)
		if err != nil {
			log.Printf("[ERROR] - While reaching the API.\n%v\n", err.Error())
			return
		}
		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Printf("[ERROR] - While reading response from the API.\n%v\n", err.Error())
			return
		}
		if api == "artists" {
			err = json.Unmarshal(data, &artistTemp)
		} else {
			err = json.Unmarshal(data, &locTemp)
		}
		if err != nil {
			log.Printf("[WARNING] - Ignoring some useless JSON data from the API.\n", err)
		}
	}
	// Deep copy of tmpBandsData to avoid storing it on memory
	// Since it's a local variable it should only exists in the function's scope.
	self.AllBands = make([]BandInfo, len(artistTemp))
	copy(self.AllBands, artistTemp)
	for index, band := range locTemp["index"] {
		formatedRelation := map[string][]string{}
		for place, date := range band.Relations {
			formatedPlace := strings.Title(strings.ReplaceAll(strings.ReplaceAll(place, "-", " - "), "_", " "))
			formatedRelation[formatedPlace] = date
		}
		self.AllBands[index].Relations = formatedRelation
	}
	// Shallow copy of AllBands since it's not a function scoped variable.
	self.DisplayBands = self.AllBands
	log.Printf("[INFO] - Data succesfully extracted!\n")
}

func AtoiSlice(seq []string) []int {
	result := make([]int, len(seq))
	for index, el := range seq {
		number, err := strconv.Atoi(el)
		if err != nil {
			log.Printf("[WARNING] - Could not convert string \"%v\" to int.", el)
			return result
		}
		result[index] = number
	}
	return result
}

func (self *ApiData) RootHandler(webpage http.ResponseWriter, request *http.Request) {
	/*
		Method of ApiData
		This is the main function of the structure, it handles an HTTP connection on the "/" path.
		If the request method is POST, it will try to perform some sorting based on the content of the
		request. Otherwise, it will simply serve the basic main HTML page with the self.DisplayBands
		displayed.
		:param webpage: the page we're writing on
		:param request: the current request
	*/
	if request.URL.Path != "/" && request.URL.Path != "" {
		ServeFile(webpage, "404.html", nil)
		return
	}

	if request.Method == "POST" {

		request.ParseForm()

		allKeys := make(map[int]bool)

		for _, element := range self.AllBands {
			allKeys[element.Id] = true
		}

		if len(request.Form["filter_startingyear"]) != 0 {
			startyear := AtoiSlice(request.Form["filter_startingyear"])[0]
			//fmt.Println(startyear)

			for _, artist := range self.AllBands {
				if startyear > artist.CreationDate {
					allKeys[artist.Id] = false
				}
			}
		}

		if len(request.Form["filter_endyear"]) != 0 {
			endyear := AtoiSlice(request.Form["filter_endyear"])[0]
			//fmt.Println(endyear)
			for _, artist := range self.AllBands {
				if endyear < artist.CreationDate {
					allKeys[artist.Id] = false
				}
			}
		}

		if len(request.Form["filter_nmembers"]) != 0 {
			sizes := AtoiSlice(request.Form["filter_nmembers"])
			//fmt.Println(sizes)
			exist := false
			//fmt.Println(sizes)
			for _, artist := range self.AllBands {
				for i := 0; i < len(sizes); i++ {
					if sizes[i] == len(artist.Members) {
						exist = true
						break
					}
				}
				if exist == false {
					allKeys[artist.Id] = false
				}
				exist = false
			}

		}
		if len(request.Form["filter_location"]) != 0 && request.Form["filter_location"][0] != "" {
			country := string(request.Form["filter_location"][0])
			country = strings.TrimSpace(strings.ToLower(strings.ReplaceAll(country, " ", "")))
			//fmt.Println(country)
			exist := false
			for _, artist := range self.AllBands {
				for location := range artist.Relations {
					// location = strings.Split(location, "- ")[1]
					l := strings.TrimSpace(strings.ToLower(strings.ReplaceAll(location, " ", "")))

					if l == country {
						//fmt.Println(l)
						exist = true
						break
					}
				}
				if !exist {
					allKeys[artist.Id] = false
				}
				exist = false
			}

		}

		if len(request.Form["filter_FAsyear"]) != 0 {
			startyear := AtoiSlice(request.Form["filter_FAsyear"])[0]
			//fmt.Println(startyear)
			for _, artist := range self.AllBands {
				date := strings.Split(artist.FirstAlbum, "-") // ["12", "02", "2020"]
				year, _ := strconv.Atoi(date[2])
				if startyear > year {
					allKeys[artist.Id] = false
				}
			}

		}
		if len(request.Form["filter_FAeyear"]) != 0 {
			endyear := AtoiSlice(request.Form["filter_FAeyear"])[0]
			//fmt.Println(endyear)
			for _, artist := range self.AllBands {
				date := strings.Split(artist.FirstAlbum, "-") // ["12", "02", "2020"]
				year, _ := strconv.Atoi(date[2])
				if endyear < year {
					allKeys[artist.Id] = false
				}
			}
		}

		for _, element := range self.AllBands {
			if allKeys[element.Id] {
				self.filterBands = append(self.filterBands, element)
			}
		}

		self.filterBands = KeepOnlyDuplicates(self.filterBands, self.AllBands)

		self.DisplayBands = self.filterBands

		//fmt.Println(self.filterBands)
		self.filterBands = []BandInfo{}

		//fmt.Println(".......................")

		if len(request.Form["input-search"]) != 0 {
			self.DisplayBands = []BandInfo{}
			data := strings.Split(request.Form["input-search"][0], " (")
			searchingFor := strings.ToLower(data[0])

			if len(data) > 1 {
				sort := strings.Trim(data[1], ")")
				for _, artist := range self.AllBands {
					if (strings.Contains(strings.ToLower(artist.Name), searchingFor) && sort == "Band/Artist") ||
						(strings.Contains(strconv.Itoa(artist.CreationDate), searchingFor) && sort == "Creation Date") ||
						(strings.Contains(strings.ToLower(artist.FirstAlbum), searchingFor) && sort == "First Album") {
						self.DisplayBands = append(self.DisplayBands, artist)
					} else {
						for _, member := range artist.Members {
							if strings.Contains(strings.ToLower(member), searchingFor) && sort == "Member" {
								self.DisplayBands = append(self.DisplayBands, artist)
								break
							}
						}
					}
					for _, band := range self.AllBands {
						for location := range band.Relations {
							llocation := strings.ReplaceAll(strings.ToLower(location), " ", "")
							if strings.Contains(llocation, strings.ReplaceAll(searchingFor, " ", "")) && sort == "Location" {
								self.DisplayBands = append(self.DisplayBands, band)
								break
							}
						}
					}
				}
			} else {
				for _, artist := range self.AllBands {
					if strings.Contains(strings.ToLower(artist.Name), searchingFor) ||
						strings.Contains(strconv.Itoa(artist.CreationDate), searchingFor) ||
						strings.Contains(strings.ToLower(artist.FirstAlbum), searchingFor) {
						self.DisplayBands = append(self.DisplayBands, artist)
					} else {
						for _, member := range artist.Members {
							if strings.Contains(strings.ToLower(member), searchingFor) {
								self.DisplayBands = append(self.DisplayBands, artist)
								break
							}
						}

					}
					for _, band := range self.AllBands {
						for location := range band.Relations {
							llocation := strings.ReplaceAll(strings.ToLower(location), " ", "")
							if strings.Contains(llocation, searchingFor) {
								self.DisplayBands = append(self.DisplayBands, band)

								break
							}
						}
					}
				}
			}

		}
		self.DisplayBands = KeepOnlyDuplicates(self.DisplayBands, self.AllBands)

	} else {
		self.DisplayBands = self.AllBands
	}

	//controller.ServeFile(webpage, "index.html", self)
	_, err := os.Stat("./view/html/index.html")
	p500, _ := template.ParseFiles("./view/html/500.html")

	if os.IsNotExist(err) {
		log.Println("[ERROR] - File 'index.html' does not exist or is not accessible.")

		webpage.WriteHeader(http.StatusInternalServerError)
		p500.ExecuteTemplate(webpage, "500.html", nil)
	} else {
		content, err := template.ParseFiles("./view/html/index.html")
		if err != nil {
			log.Printf("[ERROR] - Failed to parse 'index.html': %v\n", err.Error())

			webpage.WriteHeader(http.StatusInternalServerError)
			p500.ExecuteTemplate(webpage, "500.html", nil)
		} else {
			err = content.ExecuteTemplate(webpage, "index.html", self)
			if err != nil {
				log.Printf("[ERROR] - Failed to execute 'index.html' template: %v\n", err.Error())

				webpage.WriteHeader(http.StatusInternalServerError)
				p500.ExecuteTemplate(webpage, "500.html", nil)
			}
		}
	}
}




