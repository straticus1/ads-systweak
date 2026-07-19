package cmd

import "testing"

func TestParseSWVers(t *testing.T) {
	facts := parseSWVers("ProductName:\tmacOS\nProductVersion:\t15.7.4\nBuildVersion:\t24G508\n")
	if facts["ProductVersion"] != "15.7.4" || facts["BuildVersion"] != "24G508" {
		t.Fatalf("facts = %#v", facts)
	}
}
