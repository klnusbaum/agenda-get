package sites

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestSimpleSites(t *testing.T) {
	tests := []struct {
		site  SimpleSite
		today time.Time
		mocks map[string]string
	}{
		{
			site: Sacramento(),
			mocks: map[string]string{
				"http://sacramento.granicus.com/viewpublisher.php?view_id=34":             "sacramento.html",
				"http://sacramento.granicus.com/AgendaViewer.php?view_id=34&clip_id=4665": "testok",
			},
		},
		{
			site: Oakland(),
			mocks: map[string]string{
				"https://www.oaklandca.gov/boards-commissions/planning-commission/meetings":                        "oakland.html",
				"https://cao-94612.s3.amazonaws.com/documents/August-5-2020-Planning-Commission-Agenda-Online.pdf": "testok",
			},
		},
		{
			site: Bakersfield(),
			mocks: map[string]string{
				"https://bakersfield.novusagenda.com/AgendaPublic/?MeetingType=6":                                                    "bakersfield.html",
				"https://bakersfield.novusagenda.com/AgendaPublic/MeetingView.aspx?MeetingID=549&MinutesMeetingID=-1&doctype=Agenda": "testok",
			},
		},
		{
			site:  Fresno(),
			today: time.Date(2020, time.August, 3, 0, 0, 0, 0, time.UTC),
			mocks: map[string]string{
				"https://fresno.legistar.com/DepartmentDetail.aspx?ID=24452&GUID=26F8DAF5-AC08-46BE-A9E4-EC0C6DDC0F66&Search=": "fresno.html",
				"https://fresno.legistar.com/View.ashx?M=A&ID=749663&GUID=72538FC3-A2B5-4017-BC5A-39FE482D612E":                "testok",
			},
		},
		{
			site: SanFrancisco(),
			mocks: map[string]string{
				"https://sfplanning.org/hearings-cpc":                                         "sanfrancisco.html",
				"https://sfplanning.org/sites/default/files/agendas/2020-07/20200730_cal.pdf": "testok",
			},
		},
		{
			site: Pasadena(),
			mocks: map[string]string{
				"https://www.cityofpasadena.net/commissions/planning-commission/":                                                  "pasadena.html",
				"https://www.cityofpasadena.net/commissions/wp-content/uploads/sites/31/2020-07-22-Planning-Commission-Agenda.pdf": "testok",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.site.Entity, func(t *testing.T) {
			client := testClient{
				responses: tt.mocks,
			}
			fetcher := NewDefaultFetcher(client, tt.today)
			agenda, err := fetcher.Simple(context.Background(), tt.site)
			require.NoError(t, err)
			content, err := ioutil.ReadAll(agenda.Content)
			require.NoError(t, err)
			require.Equal(t, "OK\n", string(content))
		})
	}
}

type testClient struct {
	responses map[string]string
}

func (c testClient) Do(req *http.Request) (*http.Response, error) {
	filename, ok := c.responses[req.URL.String()]
	if !ok {
		return nil, fmt.Errorf("no response for url %s", req.URL.String())
	}
	filepath := path.Join("testdata", filename)
	testdata, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("open %s: %s", filepath, err)
	}
	return &http.Response{
		Body:       testdata,
		Request:    req,
		StatusCode: 200,
	}, nil
}
