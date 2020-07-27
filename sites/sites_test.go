package sites

import (
	"os"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/require"
)

func TestOakland(t *testing.T) {
	page, err := os.Open("testdata/oakland.html")
	require.NoError(t, err)
	doc, err := goquery.NewDocumentFromReader(page)
	require.NoError(t, err)

	latestMeeting, found := doc.
		Find("#meetings").
		Find("tbody").
		Find("tr").
		First().
		Find("td").
		Eq(4).
		Find("a").
		Attr("href")

	require.True(t, found)
	require.Equal(t, latestMeeting, "https://cao-94612.s3.amazonaws.com/documents/August-5-2020-Planning-Commission-Agenda-Online.pdf")
}
