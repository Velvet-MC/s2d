package translate

import "testing"

func TestJavaStateCoverageReportsEmbeddedStates(t *testing.T) {
	report, err := JavaStateCoverage()
	if err != nil {
		t.Fatalf("JavaStateCoverage: %v", err)
	}
	if report.TotalStates == 0 {
		t.Fatal("TotalStates = 0, want embedded Java states to be enumerated")
	}
	if report.Unknown.Total != 0 {
		t.Fatalf("embedded Java state coverage has %d unknown cells across %d states; first missing: %s",
			report.Unknown.Total, len(report.Unknown.Counts), report.TopUnknowns(1)[0].State)
	}
}
