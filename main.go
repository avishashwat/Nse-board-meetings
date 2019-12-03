package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode"
)

type Meeting struct {
	Symbol           string
	CompanyName      string
	ISIN             string
	Ind              string
	Purpose          string
	BoardMeetingDate string
	DisplayDate      string
	seqId            string
	Details          string
}

type MeetingsMapResults struct {
	Purposes          string
	BoardMeetingDates string
	Detail            string
}
type MeetingsMapKeys struct {
	Symbol  string
	Purpose string
}

type MeetingsAggregatePage struct {
	Symbol string
	Meets  map[MeetingsMapKeys]MeetingsMapResults
}

func getData() string {
	response, err := http.Get("https://www.nseindia.com/corporates/corpInfo/equities/getBoardMeetings.jsp?period=Latest%20Announced")
	if err != nil {
		fmt.Println(err.Error())
	}
	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println(err.Error())
	}
	return string(responseData)
}

func formatData(string_body string) []string {
	tempString := strings.TrimFunc(strings.TrimSpace(string_body), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	return strings.Split(tempString[(strings.Index(tempString, "["))+2:], "},{")
}

func createAggregateMeetingPage(allMeetings []Meeting) MeetingsAggregatePage {
	meetingsMap := make(map[MeetingsMapKeys]MeetingsMapResults)
	for i := 0; i < len(allMeetings); i++ {
		meetingsMap[MeetingsMapKeys{allMeetings[i].Symbol, allMeetings[i].Purpose}] = MeetingsMapResults{allMeetings[i].Purpose, allMeetings[i].BoardMeetingDate, allMeetings[i].Details}
	}
	return MeetingsAggregatePage{Symbol: "NSE BOARD MEETINGS", Meets: meetingsMap}
}

func createSliceOfMeetings(meeting []string) []Meeting {
	sprtr := "\","
	totalAttributes := 9
	temps := []string{}
	var allMeeting []Meeting

	for i := 0; i < len(meeting); i++ {
		for index := 0; index < len(strings.Split(meeting[0], sprtr)); index++ {
			temp := strings.SplitN(strings.Split(meeting[i], sprtr)[index], ":", 2)[1][1:]
			temps = append(temps, temp)
		}
		arrayOfAttributes := temps[(i * totalAttributes) : (i*totalAttributes)+totalAttributes]
		if strings.HasSuffix(arrayOfAttributes[8], "\"") {
			strlen := len(arrayOfAttributes[8])
			arrayOfAttributes[8] = (arrayOfAttributes[8])[:(strlen - 1)]
		}
		indiviualMeetingStruct := Meeting{arrayOfAttributes[0], arrayOfAttributes[1], arrayOfAttributes[2], arrayOfAttributes[3], arrayOfAttributes[4], arrayOfAttributes[5], arrayOfAttributes[6], arrayOfAttributes[7], arrayOfAttributes[8]}
		allMeeting = append(allMeeting, indiviualMeetingStruct)
	}
	return allMeeting
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	var consolidatedMeetingsData []string
	var sliceOfMeetings []Meeting
	//Getting the data
	apiResponse := getData()

	//Formatting the data
	consolidatedMeetingsData = formatData(apiResponse)

	//Creating a slice of Meetings
	sliceOfMeetings = createSliceOfMeetings(consolidatedMeetingsData)

	//Mapping for Meetings Struct
	p := createAggregateMeetingPage(sliceOfMeetings)
	t, _ := template.ParseFiles("boardmeetings.html")
	t.Execute(w, p)
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":8000", nil)
}
