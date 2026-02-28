package adapters

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

type homeCategoryDefinition struct {
	id            int64
	name          string
	subcategories []string
}

type homeSectionAgeView struct {
	Section     domain.HomeCategorySection
	CompactAge  string
	DetailedAge string
}

var defaultHomeCategoryDefinitions = []homeCategoryDefinition{
	{
		id:   domain.CategoryHousing,
		name: "housing",
		subcategories: []string{
			"apts/housing",
			"housing wanted",
			"rooms/shared",
			"sublets/temporary",
		},
	},
	{
		id:   domain.CategoryForSale,
		name: "for sale",
		subcategories: []string{
			"art+shop",
			"barter",
			"bikes",
			"books",
			"cars",
			"cds/dvds",
			"clothes+acc",
			"computers",
			"electronics",
			"free",
			"furniture",
			"games",
			"household",
			"tickets",
			"wanted",
			"general",
		},
	},
	{
		id:   domain.CategoryJobsOffCampus,
		name: "jobs off-campus",
		subcategories: []string{
			"post a job",
		},
	},
	{
		id:   domain.CategoryPersonals,
		name: "personals",
		subcategories: []string{
			"friendship",
			"girl wants girl",
			"girl wants guy",
			"guy wants girl",
			"guy wants guy",
			"general",
		},
	},
	{
		id:   domain.CategoryCampusJob,
		name: "campus job",
		subcategories: []string{
			"admin",
			"research",
			"teaching",
			"tutoring",
			"general",
		},
	},
	{
		id:   domain.CategoryCommunity,
		name: "community",
		subcategories: []string{
			"activities",
			"childcare",
			"classes",
			"lost+found",
			"news+views",
			"rideshare",
			"volunteers",
			"general",
		},
	},
	{
		id:   domain.CategoryServices,
		name: "services",
		subcategories: []string{
			"computer",
			"tutoring",
			"general",
		},
	},
}

func renderHomeOverviewAndRecent(w io.Writer, posts []domain.Post, featuredPosts []domain.Post, sections []domain.HomeCategorySection, now time.Time, width int) error {
	leftWidth, rightWidth := calculateStripWidths(width)
	normalizedSections := normalizeHomeSections(sections)
	normalizedSections = fillMissingSectionTimes(normalizedSections, posts)
	sectionViews := buildHomeSectionAgeViews(normalizedSections, now)
	leftRows := renderHomeSidebarRows(leftWidth, sectionViews)
	rightRows := renderHomeRecentAndFeaturedRows(posts, featuredPosts, now, rightWidth)

	totalRows := len(rightRows)
	if len(leftRows) > totalRows {
		totalRows = len(leftRows)
	}

	for i := 0; i < totalRows; i++ {
		left := strings.Repeat(" ", leftWidth)
		right := strings.Repeat(" ", rightWidth)
		if i < len(leftRows) {
			left = leftRows[i]
		}
		if i < len(rightRows) {
			right = rightRows[i]
		}
		if _, err := fmt.Fprintln(w, left+strings.Repeat(" ", homeStripGap)+right); err != nil {
			return err
		}
	}
	return nil
}

func renderHomeSidebarRows(width int, sectionViews []homeSectionAgeView) []string {
	rows := make([]string, 0, 48)
	rows = append(rows, renderHomeOverviewRows(width, sectionViews)...)
	rows = append(rows, strings.Repeat(" ", width))
	rows = append(rows, renderHomeCategoryDetailsRows(width, sectionViews)...)
	return rows
}

func renderHomeOverviewRows(width int, sectionViews []homeSectionAgeView) []string {
	rows := []string{
		ansiHeader + centerText("overview", width) + ansiReset,
	}
	for _, sectionView := range sectionViews {
		rows = append(rows, renderOverviewRow(sectionView.Section.CategoryName, sectionView.CompactAge, width))
	}
	return rows
}

func renderOverviewRow(label, age string, width int) string {
	if width <= 0 {
		return ""
	}

	label = strings.TrimSpace(label)
	age = strings.TrimSpace(age)
	labelLen := len([]rune(label))
	ageLen := len([]rune(age))

	minGap := 1
	if labelLen+minGap+ageLen > width {
		available := width - ageLen - minGap
		if available < 1 {
			return fitText(label+" "+age, width)
		}
		label = fitText(label, available)
		labelLen = len([]rune(label))
	}

	gap := width - labelLen - ageLen
	if gap < 1 {
		gap = 1
	}

	return ansiBlue + label + ansiReset + strings.Repeat(" ", gap) + ansiMagenta + age + ansiReset
}

func renderHomeCategoryDetailsRows(width int, sectionViews []homeSectionAgeView) []string {
	rows := make([]string, 0, len(sectionViews)*8)

	for idx, sectionView := range sectionViews {
		if idx > 0 {
			rows = append(rows, strings.Repeat(" ", width))
		}

		rows = append(rows, styleCentered(sectionView.Section.CategoryName, width, ansiBlue))
		rows = append(rows, centerText("("+sectionView.DetailedAge+")", width))
		rows = append(rows, renderSubcategoryRows(sectionView.Section.SubcategoryNames, width)...)
	}

	return rows
}

func buildHomeSectionAgeViews(sections []domain.HomeCategorySection, now time.Time) []homeSectionAgeView {
	views := make([]homeSectionAgeView, 0, len(sections))
	for _, section := range sections {
		views = append(views, homeSectionAgeView{
			Section:     section,
			CompactAge:  formatCompactAge(section.LastPostedAt, now),
			DetailedAge: formatDetailedCategoryAge(section.LastPostedAt, now),
		})
	}
	return views
}

func renderSubcategoryRows(subcategories []string, width int) []string {
	if len(subcategories) == 0 {
		return nil
	}

	if len(subcategories) <= 4 {
		rows := make([]string, 0, len(subcategories))
		for _, name := range subcategories {
			rows = append(rows, styleLeft(name, width, ansiBlue))
		}
		return rows
	}

	columnGap := 2
	columnWidth := (width - columnGap) / 2
	if columnWidth < 8 {
		rows := make([]string, 0, len(subcategories))
		for _, name := range subcategories {
			rows = append(rows, styleLeft(name, width, ansiBlue))
		}
		return rows
	}

	leftCount := (len(subcategories) + 1) / 2
	rows := make([]string, 0, leftCount)
	for i := 0; i < leftCount; i++ {
		left := fitText(subcategories[i], columnWidth)
		right := ""
		if rightIndex := i + leftCount; rightIndex < len(subcategories) {
			right = fitText(subcategories[rightIndex], columnWidth)
		}

		line := ansiBlue + left + ansiReset + strings.Repeat(" ", columnGap)
		if right != "" {
			line += ansiBlue + right + ansiReset
		} else {
			line += strings.Repeat(" ", columnWidth)
		}
		rows = append(rows, line)
	}
	return rows
}

func calculateStripWidths(totalWidth int) (calloutWidth, rightWidth int) {
	calloutWidth = homeCalloutWidth
	if totalWidth <= 0 {
		return calloutWidth, 0
	}

	minCallout := 18
	maxCallout := totalWidth / 2
	if maxCallout < minCallout {
		maxCallout = minCallout
	}
	if calloutWidth > maxCallout {
		calloutWidth = maxCallout
	}
	if calloutWidth < minCallout {
		calloutWidth = minCallout
	}

	rightWidth = totalWidth - calloutWidth - homeStripGap
	if rightWidth < 1 {
		rightWidth = 1
	}
	return calloutWidth, rightWidth
}

func renderHomeCalloutRows(width int) []string {
	return []string{
		styleCentered("post to classifieds", width, ansiBlue),
		styleCentered("@stanford.edu required", width, ansiGray),
		strings.Repeat(" ", width),
		styleCentered("post a job", width, ansiBlue),
		styleCentered("post housing", width, ansiBlue),
		styleCentered("post a car", width, ansiBlue),
		strings.Repeat(" ", width),
		styleCentered("open for all emails", width, ansiGray),
	}
}

func styleCentered(text string, width int, color string) string {
	cell := centerText(text, width)
	if strings.TrimSpace(text) == "" || color == "" {
		return cell
	}
	return color + cell + ansiReset
}

func centerText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	trimmed := strings.TrimSpace(text)
	runes := []rune(trimmed)
	if len(runes) > width {
		return fitText(trimmed, width)
	}
	padding := width - len(runes)
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + trimmed + strings.Repeat(" ", right)
}

func styleLeft(text string, width int, color string) string {
	cell := fitText(strings.TrimSpace(text), width)
	if strings.TrimSpace(text) == "" || color == "" {
		return cell
	}
	return color + cell + ansiReset
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func normalizeHomeSections(sections []domain.HomeCategorySection) []domain.HomeCategorySection {
	byID := make(map[int64]domain.HomeCategorySection, len(sections))
	for _, section := range sections {
		normalized := section
		normalized.CategoryName = strings.TrimSpace(section.CategoryName)
		normalized.SubcategoryNames = normalizeSubcategoryNames(section.SubcategoryNames)
		byID[section.CategoryID] = normalized
	}

	out := make([]domain.HomeCategorySection, 0, len(defaultHomeCategoryDefinitions))
	for _, def := range defaultHomeCategoryDefinitions {
		section, ok := byID[def.id]
		if !ok {
			out = append(out, domain.HomeCategorySection{
				CategoryID:       def.id,
				CategoryName:     def.name,
				SubcategoryNames: append([]string(nil), def.subcategories...),
			})
			continue
		}

		section.CategoryName = canonicalCategoryName(section.CategoryID, section.CategoryName)
		if len(section.SubcategoryNames) == 0 {
			section.SubcategoryNames = append([]string(nil), def.subcategories...)
		}
		out = append(out, section)
	}
	return out
}

func normalizeSubcategoryNames(names []string) []string {
	cleaned := make([]string, 0, len(names))
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		cleaned = append(cleaned, trimmed)
	}
	return cleaned
}

func canonicalCategoryName(id int64, fromDB string) string {
	for _, def := range defaultHomeCategoryDefinitions {
		if def.id == id {
			return def.name
		}
	}
	trimmed := strings.TrimSpace(fromDB)
	if trimmed == "" {
		return "general"
	}
	return trimmed
}

func formatCompactAge(from, now time.Time) string {
	if from.IsZero() {
		return "no active posts"
	}
	return strings.TrimPrefix(formatRelativeTime(from, now), "about ")
}

func formatDetailedCategoryAge(from, now time.Time) string {
	if from.IsZero() {
		return "no active posts"
	}
	return formatRelativeTime(from, now)
}

func fillMissingSectionTimes(sections []domain.HomeCategorySection, posts []domain.Post) []domain.HomeCategorySection {
	if len(sections) == 0 || len(posts) == 0 {
		return sections
	}

	latestByCategory := make(map[int64]time.Time, len(sections))
	for _, post := range posts {
		if post.Status != domain.PostStatusActive {
			continue
		}
		t := postTimestamp(post)
		if t.IsZero() {
			continue
		}
		if existing, ok := latestByCategory[post.CategoryID]; !ok || t.After(existing) {
			latestByCategory[post.CategoryID] = t
		}
	}

	out := make([]domain.HomeCategorySection, 0, len(sections))
	for _, section := range sections {
		updated := section
		if updated.LastPostedAt.IsZero() {
			if latest, ok := latestByCategory[updated.CategoryID]; ok {
				updated.LastPostedAt = latest
			}
		}
		out = append(out, updated)
	}
	return out
}
