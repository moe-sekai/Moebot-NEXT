package assets

import "testing"

func TestChartSourceURLExpandsTemplate(t *testing.T) {
	got := ChartSourceURL("https://charts.example.test/{id}/{difficulty}.svg", 739, "master")
	if got != "https://charts.example.test/739/master.svg" {
		t.Fatalf("ChartSourceURL = %q", got)
	}
}

func TestChartSourceURLDefaultsDifficultyAndTemplate(t *testing.T) {
	got := ChartSourceURL("", 739, "")
	want := "https://charts-new.unipjsk.com/moe/svg/739/master.svg"
	if got != want {
		t.Fatalf("ChartSourceURL = %q, want %q", got, want)
	}
}
