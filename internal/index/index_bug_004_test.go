package index

import "testing"

func sampleSegments() []SegmentMeta {
	return []SegmentMeta{{ID: "a", Docs: 30}, {ID: "b", Docs: 10}, {ID: "c", Docs: 20}}
}

func TestFilterSegmentsKeepsInput(t *testing.T) {
	in := sampleSegments()
	got := FilterSegments(in, func(m SegmentMeta) bool { return m.Docs >= 20 })
	if len(got) != 2 || len(in) != 3 || in[1].ID != "b" {
		t.Fatalf("filter changed caller slice: got=%#v input=%#v", got, in)
	}
}

func TestSortSegmentsKeepsInputOrder(t *testing.T) {
	in := sampleSegments()
	got := SortSegments(in)
	if got[0].ID != "b" || in[0].ID != "a" {
		t.Fatalf("sort changed caller slice: got=%#v input=%#v", got, in)
	}
}

func TestCopySegmentMetaIndependent(t *testing.T) {
	in := sampleSegments()
	got := CopySegmentMeta(in)
	got[0].ID = "changed"
	if in[0].ID == "changed" {
		t.Fatal("copy still aliases input")
	}
}
