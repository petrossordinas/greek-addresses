package service

import "testing"

func TestFilterStreetsByNumber(t *testing.T) {
	tests := []struct {
		name         string
		streets      []Street
		streetNumber string
		wantNames    []string // Name of each expected result, in order
	}{
		{
			name:         "no ranges matches any number",
			streets:      []Street{{ID: 1, Name: "ΠΑΝΕΠΙΣΤΗΜΙΟΥ", Ranges: ""}},
			streetNumber: "42",
			wantNames:    []string{"ΠΑΝΕΠΙΣΤΗΜΙΟΥ 42"},
		},
		{
			name: "odd number within the odd range matches",
			streets: []Street{{ID: 1, Name: "ΑΓΙΟΥ ΚΩΝΣΤΑΝΤΙΝΟΥ",
				Ranges: `{"odd": [{"from": "1", "to": "23"}], "even": [{"from": "2", "to": "20"}]}`}},
			streetNumber: "5",
			wantNames:    []string{"ΑΓΙΟΥ ΚΩΝΣΤΑΝΤΙΝΟΥ 5"},
		},
		{
			name: "even number outside the even range is excluded",
			streets: []Street{{ID: 1, Name: "ΑΓΙΟΥ ΚΩΝΣΤΑΝΤΙΝΟΥ",
				Ranges: `{"odd": [{"from": "1", "to": "23"}], "even": [{"from": "2", "to": "20"}]}`}},
			streetNumber: "22",
			wantNames:    nil,
		},
		{
			name:         "odd number with no odd range defined is excluded",
			streets:      []Street{{ID: 1, Name: "ΜΟΝΟ ΖΥΓΑ", Ranges: `{"even": [{"from": "2", "to": "10"}]}`}},
			streetNumber: "3",
			wantNames:    nil,
		},
		{
			name:         "an empty 'to' defaults to 999",
			streets:      []Street{{ID: 1, Name: "ΑΝΟΙΧΤΟ ΑΚΡΟ", Ranges: `{"odd": [{"from": "1", "to": ""}]}`}},
			streetNumber: "501",
			wantNames:    []string{"ΑΝΟΙΧΤΟ ΑΚΡΟ 501"},
		},
		{
			name:         "a single-byte trailing letter is stripped for matching but kept in the name",
			streets:      []Street{{ID: 1, Name: "ΜΕ ΓΡΑΜΜΑ", Ranges: `{"even": [{"from": "2", "to": "20"}]}`}},
			streetNumber: "12A",
			wantNames:    []string{"ΜΕ ΓΡΑΜΜΑ 12A"},
		},
		{
			name:         "malformed ranges JSON excludes the street instead of panicking",
			streets:      []Street{{ID: 1, Name: "ΧΑΛΑΣΜΕΝΟ", Ranges: `not json`}},
			streetNumber: "4",
			wantNames:    nil,
		},
		{
			name: "only the matching street among several is returned",
			streets: []Street{
				{ID: 1, Name: "Α", Ranges: `{"odd": [{"from": "1", "to": "9"}]}`},
				{ID: 2, Name: "Β", Ranges: `{"even": [{"from": "2", "to": "9"}]}`},
			},
			streetNumber: "3",
			wantNames:    []string{"Α 3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterStreetsByNumber(tt.streets, tt.streetNumber)
			if len(got) != len(tt.wantNames) {
				t.Fatalf("filterStreetsByNumber() returned %d streets, want %d: %+v", len(got), len(tt.wantNames), got)
			}
			for i, name := range tt.wantNames {
				if got[i].Name != name {
					t.Errorf("result[%d].Name = %q, want %q", i, got[i].Name, name)
				}
			}
		})
	}
}

func TestLikePattern(t *testing.T) {
	tests := []struct {
		qry  string
		want string
	}{
		{"ΑΘΗΝΑ", "ΑΘΗΝΑ%"},
		{"%ΑΘΗΝΑ%", "%ΑΘΗΝΑ%"},
		{"%ΑΘΗΝΑ", "%ΑΘΗΝΑ"},
		{"ΑΘΗΝΑ%", "ΑΘΗΝΑ%"},
	}

	for _, tt := range tests {
		if got := likePattern(tt.qry); got != tt.want {
			t.Errorf("likePattern(%q) = %q, want %q", tt.qry, got, tt.want)
		}
	}
}
