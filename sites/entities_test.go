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

func TestSites(t *testing.T) {
	tests := []struct {
		entity string
		site   Site
		today  time.Time
		mocks  map[string]string
	}{
		{
			entity: "oakland",
			site:   Oakland(),
			mocks: map[string]string{
				"https://www.oaklandca.gov/boards-commissions/planning-commission/meetings":                        "oakland.html",
				"https://cao-94612.s3.amazonaws.com/documents/August-5-2020-Planning-Commission-Agenda-Online.pdf": "testok",
			},
		},
		{
			entity: "bakersfield",
			site:   Bakersfield(),
			mocks: map[string]string{
				"https://bakersfield.novusagenda.com/AgendaPublic/?MeetingType=6":                                                    "bakersfield.html",
				"https://bakersfield.novusagenda.com/AgendaPublic/MeetingView.aspx?MeetingID=549&MinutesMeetingID=-1&doctype=Agenda": "testok",
			},
		},
		{
			entity: "fresno",
			site:   Fresno(),
			today:  time.Date(2020, time.August, 3, 0, 0, 0, 0, time.UTC),
			mocks: map[string]string{
				"https://fresno.legistar.com/DepartmentDetail.aspx?ID=24452&GUID=26F8DAF5-AC08-46BE-A9E4-EC0C6DDC0F66&Search=": "fresno.html",
				"https://fresno.legistar.com/View.ashx?M=A&ID=749663&GUID=72538FC3-A2B5-4017-BC5A-39FE482D612E":                "testok",
			},
		},
		{
			entity: "sanfrancisco",
			site:   SanFrancisco(),
			mocks: map[string]string{
				"https://sfplanning.org/hearings-cpc":                                         "sanfrancisco.html",
				"https://sfplanning.org/sites/default/files/agendas/2020-07/20200730_cal.pdf": "testok",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.entity, func(t *testing.T) {
			client := testClient{
				responses: tt.mocks,
			}
			agenda, err := tt.site.Get(context.Background(), client, tt.today)
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
		Body:    testdata,
		Request: req,
	}, nil
}
