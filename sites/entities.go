package sites

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var singleQuoteMatcher = regexp.MustCompile("'(.+?)'")

func Oakland() SimpleSite {
	return SimpleSite{
		"oakland",
		"https://www.oaklandca.gov/boards-commissions/planning-commission/meetings",
		func(doc *goquery.Document, today time.Time) (string, error) {
			link, found := doc.
				Find("#meetings").
				Find("tbody").
				Find("tr").
				First().
				Find("td").
				Eq(4).
				Find("a").
				Attr("href")
			if !found {
				return "", errors.New("couldn't find href attribute")
			}
			return link, nil
		},
	}
}

func Bakersfield() SimpleSite {
	return SimpleSite{
		"bakersfield",
		"https://bakersfield.novusagenda.com/AgendaPublic/?MeetingType=6",
		func(doc *goquery.Document, today time.Time) (string, error) {
			js, found := doc.
				Find("#ctl00_ContentPlaceHolder1_SearchAgendasMeetings_radGridMeetings_ctl00").
				Find("tbody").
				Find("tr").
				First().
				Find("td").
				Eq(4).
				Find("a").
				Attr("onclick")
			if !found {
				return "", errors.New("couldn't find onclick attribute")
			}
			matches := singleQuoteMatcher.FindStringSubmatch(js)
			if len(matches) < 2 {
				return "", fmt.Errorf("couldn't parse url from javascript: %s", js)
			}
			return "https://bakersfield.novusagenda.com/AgendaPublic/" + matches[1], nil
		},
	}
}

func Fresno() SimpleSite {
	return SimpleSite{
		"fresno",
		"https://fresno.legistar.com/DepartmentDetail.aspx?ID=24452&GUID=26F8DAF5-AC08-46BE-A9E4-EC0C6DDC0F66&Search=",
		func(doc *goquery.Document, today time.Time) (string, error) {
			meetings := doc.
				Find("#ctl00_ContentPlaceHolder1_gridCalendar_ctl00").
				Find("tbody").
				Find("tr")
			if meetings.Length() == 0 {
				return "", errors.New("no meetings this month")
			}

			link, found := meetings.
				FilterFunction(func(i int, sel *goquery.Selection) bool {
					date, err := time.Parse("1/2/2006", sel.Find("td").Eq(0).Text())
					if err != nil {
						return false
					}
					return date.After(today)
				}).
				Last().
				Find("td").
				Eq(5).
				Find("a").
				Attr("href")
			if !found {
				return "", errors.New("couldn't find href attribute")
			}

			return "https://fresno.legistar.com/" + link, nil
		},
	}
}

func SanFrancisco() SimpleSite {
	return SimpleSite{
		"sanfrancisco",
		"https://sfplanning.org/hearings-cpc",
		func(doc *goquery.Document, today time.Time) (string, error) {
			link, found := doc.
				Find("div.view-content").
				Find("div.views-row").
				First().
				Find("div.right").
				Find("a:contains(AGENDA)").
				Attr("href")
			if !found {
				return "", errors.New("couldn't find href attribute")
			}
			return link, nil
		},
	}
}
