package migration

import "testing"

func TestAdd(t *testing.T) {
	mp := MigrationPlan{}

	m1 := Migration{
		Name:      "m1",
		UpQuery:   "Dummy up query 1",
		DownQuery: "Dummy down query 1",
	}

	m2 := Migration{
		Name:      "m2",
		UpQuery:   "Dummy up query 2",
		DownQuery: "Dummy down query 2",
		Next:      &m1,
	}

	m3 := Migration{
		Name:      "m3",
		UpQuery:   "Dummy up query 3",
		DownQuery: "Dummy down query 3",
		Previous:  &m1,
		Next:      &m2,
	}

	mp.Add(m1)
	mp.Add(m2)
	mp.Add(m3)

	// mp1

	mp1 := mp.First

	if mp1.Name != m1.Name {
		t.Fatalf("Incorrect migration name: '%s' != '%s'.", mp1.Name, m1.Name)
	}

	if mp1.UpQuery != m1.UpQuery {
		t.Fatalf("Incorrect UpQuery: '%s' != '%s'.", mp1.UpQuery, m1.UpQuery)
	}

	if mp1.DownQuery != m1.DownQuery {
		t.Fatalf("Incorrect DownQuery: '%s' != '%s'.", mp1.DownQuery, m1.DownQuery)
	}

	if nil != mp1.Previous {
		t.Fatal("mp1.Previous is not nil.")
	}

	// mp2

	mp2 := mp1.Next

	if mp2.Name != m2.Name {
		t.Fatalf("Incorrect migration name: '%s' != '%s'.", mp2.Name, m2.Name)
	}

	if mp2.UpQuery != m2.UpQuery {
		t.Fatalf("Incorrect UpQuery: '%s' != '%s'.", mp2.UpQuery, m2.UpQuery)
	}

	if mp2.DownQuery != m2.DownQuery {
		t.Fatalf("Incorrect DownQuery: '%s' != '%s'.", mp2.DownQuery, m2.DownQuery)
	}

	if mp2.Previous != mp1 {
		t.Fatal("mp2.Previous != mp1.")
	}

	// mp3

	mp3 := mp2.Next

	if mp3.Name != m3.Name {
		t.Fatalf("Incorrect migration name: '%s' != '%s'.", mp3.Name, m3.Name)
	}

	if mp3.UpQuery != m3.UpQuery {
		t.Fatalf("Incorrect UpQuery: '%s' != '%s'.", mp3.UpQuery, m3.UpQuery)
	}

	if mp3.DownQuery != m3.DownQuery {
		t.Fatalf("Incorrect DownQuery: '%s' != '%s'.", mp3.DownQuery, m3.DownQuery)
	}

	if mp3.Previous != mp2 {
		t.Fatal("mp3.Previous != mp2.")
	}

	if nil != mp3.Next {
		t.Fatal("mp3.Next != nil.")
	}
}

func TestConcat(t *testing.T) {
	// mp1

	mp1 := MigrationPlan{}
	m1 := Migration{
		Name:      "m1",
		UpQuery:   "Dummy up query 1",
		DownQuery: "Dummy down query 1",
	}
	m2 := Migration{
		Name:      "m2",
		UpQuery:   "Dummy up query 2",
		DownQuery: "Dummy down query 2",
	}
	mp1.Add(m1).Add(m2)

	// mp2

	mp2 := MigrationPlan{}
	m3 := Migration{
		Name:      "m3",
		UpQuery:   "Dummy up query 3",
		DownQuery: "Dummy down query 3",
	}
	mp2.Add(m3)

	// mp3

	mp3 := MigrationPlan{}
	m4 := Migration{
		Name:      "m4",
		UpQuery:   "Dummy up query 4",
		DownQuery: "Dummy down query 4",
	}
	m5 := Migration{
		Name:      "m5",
		UpQuery:   "Dummy up query 5",
		DownQuery: "Dummy down query 5",
	}
	m6 := Migration{
		Name:      "m6",
		UpQuery:   "Dummy up query 6",
		DownQuery: "Dummy down query 6",
	}
	mp3.Add(m4).Add(m5).Add(m6)

	// concat

	mp1.Concat(&mp2, &mp3)

	// perform checks on concatenated data

	mc1 := mp1.First

	if mc1.Name != m1.Name {
		t.Fatalf("Mismatch between mc1.Name and m1.Name: '%s' != '%s'", mc1.Name, m1.Name)
	}
	if mc1.DownQuery != m1.DownQuery {
		t.Fatalf("Mismatch between mc1.DownQuery and m1.DownQuery: '%s' != '%s'", mc1.DownQuery, m1.DownQuery)
	}
	if mc1.UpQuery != m1.UpQuery {
		t.Fatalf("Mismatch between mc1.UpQuery and m1.UpQuery: '%s' != '%s'", mc1.UpQuery, m1.UpQuery)
	}
	if mc1.Previous != nil {
		t.Fatalf("mc1.Previous should be nil: '%s' != '%s'", mc1.UpQuery, m1.UpQuery)
	}
	if mc1.Next == m1.Next {
		t.Fatalf("mc1.Next and m1.Next reference the same data.")
	}

	mc2 := mc1.Next

	if mc2.Name != m2.Name {
		t.Fatalf("Mismatch between mc2.Name and m2.Name: '%s' != '%s'", mc2.Name, m2.Name)
	}
	if mc2.DownQuery != m2.DownQuery {
		t.Fatalf("Mismatch between mc2.DownQuery and m2.DownQuery: '%s' != '%s'", mc2.DownQuery, m2.DownQuery)
	}
	if mc2.UpQuery != m2.UpQuery {
		t.Fatalf("Mismatch between mc2.UpQuery and m2.UpQuery: '%s' != '%s'", mc2.UpQuery, m2.UpQuery)
	}
	if mc2.Previous == m2.Previous {
		t.Fatalf("mc2.Previous and m2.Previous reference the same data.")
	}
	if mc2.Next == m2.Next {
		t.Fatalf("mc2.Next and m2.Next reference the same data.")
	}

	mc3 := mc2.Next

	if mc3.Name != m3.Name {
		t.Fatalf("Mismatch between mc3.Name and m3.Name: '%s' != '%s'", mc3.Name, m3.Name)
	}
	if mc3.DownQuery != m3.DownQuery {
		t.Fatalf("Mismatch between mc3.DownQuery and m3.DownQuery: '%s' != '%s'", mc3.DownQuery, m3.DownQuery)
	}
	if mc3.UpQuery != m3.UpQuery {
		t.Fatalf("Mismatch between mc3.UpQuery and m3.UpQuery: '%s' != '%s'", mc3.UpQuery, m3.UpQuery)
	}
	if mc3.Previous == m3.Previous {
		t.Fatalf("mc3.Previous and m3.Previous reference the same data.")
	}
	if mc3.Next == m3.Next {
		t.Fatalf("mc3.Next and m3.Next reference the same data.")
	}

	mc4 := mc3.Next

	if mc4.Name != m4.Name {
		t.Fatalf("Mismatch between mc4.Name and m4.Name: '%s' != '%s'", mc4.Name, m4.Name)
	}
	if mc4.DownQuery != m4.DownQuery {
		t.Fatalf("Mismatch between mc4.DownQuery and m4.DownQuery: '%s' != '%s'", mc4.DownQuery, m4.DownQuery)
	}
	if mc4.UpQuery != m4.UpQuery {
		t.Fatalf("Mismatch between mc4.UpQuery and m4.UpQuery: '%s' != '%s'", mc4.UpQuery, m4.UpQuery)
	}
	if mc4.Previous == m4.Previous {
		t.Fatalf("mc4.Previous and m4.Previous reference the same data.")
	}
	if mc4.Next == m4.Next {
		t.Fatalf("mc4.Next and m4.Next reference the same data.")
	}

	mc5 := mc4.Next

	if mc5.Name != m5.Name {
		t.Fatalf("Mismatch between mc5.Name and m5.Name: '%s' != '%s'", mc5.Name, m5.Name)
	}
	if mc5.DownQuery != m5.DownQuery {
		t.Fatalf("Mismatch between mc5.DownQuery and m5.DownQuery: '%s' != '%s'", mc5.DownQuery, m5.DownQuery)
	}
	if mc5.UpQuery != m5.UpQuery {
		t.Fatalf("Mismatch between mc5.UpQuery and m5.UpQuery: '%s' != '%s'", mc5.UpQuery, m5.UpQuery)
	}
	if mc5.Previous == m5.Previous {
		t.Fatalf("mc5.Previous and m5.Previous reference the same data.")
	}
	if mc5.Next == m5.Next {
		t.Fatalf("mc5.Next and m5.Next reference the same data.")
	}

	mc6 := mc5.Next

	if mc6.Name != m6.Name {
		t.Fatalf("Mismatch between mc6.Name and m6.Name: '%s' != '%s'", mc6.Name, m6.Name)
	}
	if mc6.DownQuery != m6.DownQuery {
		t.Fatalf("Mismatch between mc6.DownQuery and m6.DownQuery: '%s' != '%s'", mc6.DownQuery, m6.DownQuery)
	}
	if mc6.UpQuery != m6.UpQuery {
		t.Fatalf("Mismatch between mc6.UpQuery and m6.UpQuery: '%s' != '%s'", mc6.UpQuery, m6.UpQuery)
	}
	if mc6.Previous == m6.Previous {
		t.Fatalf("mc6.Previous and m6.Previous reference the same data.")
	}
	if mc6.Next != nil {
		t.Fatalf("mc6.Next is not nil.")
	}

	// make sure original data was not transformed

	if m2.Next != nil {
		t.Fatalf("m2.Next should be nil.")
	}
	if m3.Previous != nil {
		t.Fatalf("m3.Previous should be nil.")
	}
	if m3.Next != nil {
		t.Fatalf("m3.Next should be nil.")
	}
	if m4.Previous != nil {
		t.Fatalf("m4.Previous should be nil.")
	}
	if m6.Next != nil {
		t.Fatalf("m6.Next should be nil.")
	}
}
