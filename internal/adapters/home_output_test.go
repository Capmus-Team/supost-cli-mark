package adapters

import (
	"strings"
	"testing"
	"time"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

func TestWrapStyledWords_RespectsWidth(t *testing.T) {
	words := []styledWord{
		{text: "one", color: ansiBlue},
		{text: "two", color: ansiBlue},
		{text: "three", color: ansiGray},
		{text: "four", color: ansiMagenta},
		{text: "five"},
		{text: "six"},
		{text: "seven"},
	}
	lines := wrapStyledWords(words, 10)
	if len(lines) == 0 {
		t.Fatalf("expected wrapped lines")
	}

	for _, lineWords := range lines {
		rendered := renderStyledLine(lineWords)
		plain := stripANSI(rendered)
		if got := len([]rune(plain)); got > 10 {
			t.Fatalf("line %q exceeds width: %d", plain, got)
		}
	}
}

func TestWrapStyledWords_SplitsLongWord(t *testing.T) {
	lines := wrapStyledWords([]styledWord{{text: "supercalifragilisticexpialidocious", color: ansiBlue}}, 8)
	if len(lines) < 2 {
		t.Fatalf("expected long word to be split, got %d lines", len(lines))
	}
	for _, lineWords := range lines {
		rendered := renderStyledLine(lineWords)
		plain := stripANSI(rendered)
		if got := len([]rune(plain)); got > 8 {
			t.Fatalf("line %q exceeds width: %d", plain, got)
		}
	}
}

func TestFormatDisplayEmail_StanfordDomainCollapses(t *testing.T) {
	got := formatDisplayEmail("wientjes@alumni.stanford.edu")
	if got != "@stanford.edu" {
		t.Fatalf("got %q, want %q", got, "@stanford.edu")
	}
}

func TestFormatDisplayEmail_NonStanfordUnchanged(t *testing.T) {
	got := formatDisplayEmail("person@example.com")
	if got != "" {
		t.Fatalf("got %q, want empty string", got)
	}
}

func TestSelectRecentImagePosts_ReturnsTopN(t *testing.T) {
	posts := []domain.Post{
		{ID: 130031934, HasImage: true},
		{ID: 130031933, HasImage: true},
		{ID: 130031932, HasImage: false},
		{ID: 130031931, HasImage: true},
		{ID: 130031930, HasImage: true},
		{ID: 130031929, HasImage: true},
	}

	got := selectRecentImagePosts(posts, 4)
	if len(got) != 4 {
		t.Fatalf("expected 4 image posts, got %d", len(got))
	}
	if got[0].ID != 130031934 || got[1].ID != 130031933 || got[2].ID != 130031931 || got[3].ID != 130031930 {
		t.Fatalf("unexpected IDs: %#v", got)
	}
}

func TestFormatTickerImageURL_UsesPattern(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.FixedZone("PST", -8*60*60))
	post := domain.Post{
		ID:         130031934,
		TimePosted: 1772222959,
		HasImage:   true,
	}
	got := formatTickerImageURL(post, now)
	want := "https://supost-prod.s3.amazonaws.com/posts/130031934/ticker_130031934a?1772222959"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRenderColumnRow_RespectsTotalWidth(t *testing.T) {
	line := renderColumnRow([]string{"one", "two", "three", "four"}, 10, "", 4)
	if got := len([]rune(line)); got != 46 {
		t.Fatalf("unexpected row width: got %d want %d", got, 46)
	}
}

func TestWrapColumnValue_DoesNotTruncate(t *testing.T) {
	value := "https://supost-prod.s3.amazonaws.com/posts/130031934/ticker_130031934a?1772222959"
	lines := wrapColumnValue(value, 20)
	if len(lines) < 2 {
		t.Fatalf("expected wrapped lines, got %d", len(lines))
	}

	joined := strings.Join(lines, "")
	if joined != value {
		t.Fatalf("wrapped content mismatch: got %q want %q", joined, value)
	}
	for _, line := range lines {
		if got := len([]rune(line)); got > 20 {
			t.Fatalf("line exceeds width: %d", got)
		}
	}
}

func TestRenderWrappedColumnRows_ContainsNoEllipsis(t *testing.T) {
	rows := renderWrappedColumnRows(
		[]string{"https://supost-prod.s3.amazonaws.com/posts/130031934/ticker_130031934a?1772222959"},
		25,
		"",
		1,
	)
	if len(rows) < 2 {
		t.Fatalf("expected multiple wrapped rows, got %d", len(rows))
	}
	for _, row := range rows {
		if strings.Contains(row, "…") {
			t.Fatalf("row should not contain ellipsis: %q", row)
		}
	}
}

func TestRenderHomeCalloutRows_ContainsRequestedCopy(t *testing.T) {
	rows := renderHomeCalloutRows(28)
	if len(rows) != 8 {
		t.Fatalf("expected 8 rows, got %d", len(rows))
	}

	want := []string{
		"post to classifieds",
		"@stanford.edu required",
		"post a job",
		"post housing",
		"post a car",
		"open for all emails",
	}
	plain := make([]string, 0, len(rows))
	for _, row := range rows {
		plain = append(plain, stripANSI(row))
	}
	joined := strings.Join(plain, "\n")
	for _, needle := range want {
		if !strings.Contains(joined, needle) {
			t.Fatalf("missing %q in callout rows", needle)
		}
	}
}

func TestCenterText_ProducesFixedWidth(t *testing.T) {
	got := centerText("post housing", 28)
	if len([]rune(got)) != 28 {
		t.Fatalf("expected width 28, got %d", len([]rune(got)))
	}
	if strings.TrimSpace(got) != "post housing" {
		t.Fatalf("unexpected centered content %q", got)
	}
}

func TestRenderHomeOverviewRows_ContainsRequestedCopy(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.UTC)
	sections := normalizeHomeSections(nil)
	sectionViews := buildHomeSectionAgeViews(sections, now)
	rows := renderHomeOverviewRows(28, sectionViews)
	if len(rows) != 8 {
		t.Fatalf("expected 8 overview rows, got %d", len(rows))
	}

	joined := strings.Join([]string{
		stripANSI(rows[0]),
		stripANSI(rows[1]),
		stripANSI(rows[2]),
		stripANSI(rows[3]),
		stripANSI(rows[4]),
		stripANSI(rows[5]),
		stripANSI(rows[6]),
		stripANSI(rows[7]),
	}, "\n")
	for _, needle := range []string{
		"overview",
		"housing",
		"for sale",
		"jobs",
		"personals",
		"campus job",
		"community",
		"services",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("missing %q in overview block", needle)
		}
	}
}

func TestRenderHomeCategoryDetailsRows_IncludesRequestedSubcategories(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.UTC)
	sections := normalizeHomeSections(nil)
	sectionViews := buildHomeSectionAgeViews(sections, now)
	rows := renderHomeCategoryDetailsRows(36, sectionViews)
	if len(rows) == 0 {
		t.Fatalf("expected detail rows")
	}

	joined := strings.Join(func() []string {
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, stripANSI(row))
		}
		return out
	}(), "\n")

	for _, needle := range []string{
		"housing",
		"apts/housing",
		"for sale",
		"electronics",
		"jobs off-campus",
		"post a job",
		"personals",
		"girl wants guy",
		"campus job",
		"research",
		"community",
		"rideshare",
		"services",
		"tutoring",
	} {
		if !strings.Contains(joined, needle) {
			t.Fatalf("missing %q in category detail block", needle)
		}
	}
}

func TestFormatCompactAge_ZeroTime(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.UTC)
	got := formatCompactAge(time.Time{}, now)
	if got != "no active posts" {
		t.Fatalf("got %q, want %q", got, "no active posts")
	}
}

func TestRenderRecentPostRows_RespectsWrapWidth(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.UTC)
	rows := renderRecentPostRows([]domain.Post{
		{
			ID:         1,
			Name:       "Very long listing title to ensure the rendered text wraps correctly across lines",
			Email:      "person@stanford.edu",
			HasPrice:   true,
			Price:      100,
			Status:     domain.PostStatusActive,
			TimePosted: now.Add(-2 * time.Hour).Unix(),
		},
	}, now, 20, 40)

	if len(rows) < 2 {
		t.Fatalf("expected header + body rows, got %d", len(rows))
	}

	for i := 1; i < len(rows); i++ {
		plain := stripANSI(rows[i])
		if got := len([]rune(plain)); got > 20 {
			t.Fatalf("row %d width exceeds wrap width: %d", i, got)
		}
	}
}

func TestSelectFeaturedJobPosts_FiltersActiveJobsAndOrdersNewest(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.UTC)
	posts := []domain.Post{
		{ID: 1, CategoryID: domain.CategoryHousing, Status: domain.PostStatusActive, TimePosted: now.Add(-1 * time.Hour).Unix()},
		{ID: 2, CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-4 * time.Hour).Unix()},
		{ID: 3, CategoryID: domain.CategoryJobsOffCampus, Status: 0, TimePosted: now.Add(-2 * time.Hour).Unix()},
		{ID: 4, CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-3 * time.Hour).Unix()},
		{ID: 5, CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-30 * time.Minute).Unix()},
		{ID: 6, CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-20 * time.Minute).Unix()},
	}

	got := selectFeaturedJobPosts(posts, 3)
	if len(got) != 3 {
		t.Fatalf("expected 3 posts, got %d", len(got))
	}
	if got[0].ID != 6 || got[1].ID != 5 || got[2].ID != 4 {
		t.Fatalf("unexpected post order: got ids [%d, %d, %d]", got[0].ID, got[1].ID, got[2].ID)
	}
}

func TestRenderHomeRecentAndFeaturedRows_ContainsFeaturedSection(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.UTC)
	posts := []domain.Post{
		{ID: 1, Name: "Recent housing post", CategoryID: domain.CategoryHousing, Status: domain.PostStatusActive, Email: "person@stanford.edu", TimePosted: now.Add(-10 * time.Minute).Unix()},
		{ID: 2, Name: "AI Algorithm Engineer", CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-20 * time.Minute).Unix()},
		{ID: 3, Name: "Dog Sitter During Spring Break", CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, Email: "helper@stanford.edu", TimePosted: now.Add(-30 * time.Minute).Unix()},
		{ID: 4, Name: "Software and hardware engineers", CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-40 * time.Minute).Unix()},
	}

	rows := renderHomeRecentAndFeaturedRows(posts, nil, now, 80)
	if len(rows) == 0 {
		t.Fatalf("expected rendered rows")
	}

	joined := strings.Join(func() []string {
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, stripANSI(row))
		}
		return out
	}(), "\n")
	normalized := strings.Join(strings.Fields(joined), " ")

	for _, needle := range []string{
		"recently posted",
		"featured job posts",
		"events",
		"events data placeholder",
		"Safety: If someone sends you a check, do not send them any money back. Never.",
		"SUpost is not sponsored by, endorsed by, or affiliated with Stanford University.",
		"AI Algorithm Engineer",
		"Dog Sitter During Spring Break",
		"Software and hardware engineers",
	} {
		if !strings.Contains(normalized, needle) {
			t.Fatalf("missing %q in home content rows", needle)
		}
	}
}

func TestRenderHomeRecentAndFeaturedRows_UsesExplicitFeaturedPosts(t *testing.T) {
	now := time.Date(2026, time.February, 27, 14, 25, 0, 0, time.UTC)
	posts := []domain.Post{
		{ID: 1, Name: "Recent housing post", CategoryID: domain.CategoryHousing, Status: domain.PostStatusActive, TimePosted: now.Add(-10 * time.Minute).Unix()},
	}
	featuredPosts := []domain.Post{
		{ID: 11, Name: "AI Algorithm Engineer", CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-20 * time.Minute).Unix()},
		{ID: 12, Name: "Dog Sitter During Spring Break", CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-30 * time.Minute).Unix()},
		{ID: 13, Name: "Software and hardware engineers", CategoryID: domain.CategoryJobsOffCampus, Status: domain.PostStatusActive, TimePosted: now.Add(-40 * time.Minute).Unix()},
	}

	rows := renderHomeRecentAndFeaturedRows(posts, featuredPosts, now, 80)
	joined := strings.Join(func() []string {
		out := make([]string, 0, len(rows))
		for _, row := range rows {
			out = append(out, stripANSI(row))
		}
		return out
	}(), "\n")
	normalized := strings.Join(strings.Fields(joined), " ")

	for _, needle := range []string{
		"featured job posts",
		"events",
		"events data placeholder",
		"Safety: If someone sends you a check, do not send them any money back. Never.",
		"SUpost is not sponsored by, endorsed by, or affiliated with Stanford University.",
		"AI Algorithm Engineer",
		"Dog Sitter During Spring Break",
		"Software and hardware engineers",
	} {
		if !strings.Contains(normalized, needle) {
			t.Fatalf("missing %q in explicit featured section", needle)
		}
	}
}

func stripANSI(s string) string {
	// Minimal scrubber for tests.
	out := make([]rune, 0, len(s))
	inEscape := false
	for _, r := range s {
		if r == 0x1b {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}
