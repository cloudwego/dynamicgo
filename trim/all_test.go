package trim

import "testing"

func TestFetchAndAssign(t *testing.T) {
	src := makeSampleFetch(3, 3)
	desc := makeDesc(3, 3)
	m, err := FetchAny(desc, src)
	if err != nil {
		t.Fatalf("FetchAny failed: %v", err)
	}
	dest := makeSampleAssign(3, 3)
	err = AssignAny(desc, m, dest)
	if err != nil {
		t.Fatalf("AssignAny failed: %v", err)
	}
}
