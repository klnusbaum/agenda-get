package sites

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSites(t *testing.T) {
	testData := map[string]map[string]string{
		"oakland": {

			"https://www.oaklandca.gov/boards-commissions/planning-commission/meetings":                        "oakland.html",
			"https://cao-94612.s3.amazonaws.com/documents/August-5-2020-Planning-Commission-Agenda-Online.pdf": "testok.pdf",
		},
		"bakersfield": {
			"https://bakersfield.novusagenda.com/AgendaPublic/?MeetingType=6":                                                    "bakersfield.html",
			"https://bakersfield.novusagenda.com/AgendaPublic/MeetingView.aspx?MeetingID=549&MinutesMeetingID=-1&doctype=Agenda": "testok.pdf",
		},
	}

	for _, site := range Sites {
		t.Run(site.Entity(), func(t *testing.T) {
			client := testClient{
				responses: testData[site.Entity()],
			}
			agenda, err := site.Get(context.Background(), client)
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
