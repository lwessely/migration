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
