package adapters

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Capmus-Team/supost-cli-mark/internal/domain"
)

const (
	ansiBlue    = "\033[1;34m"
	ansiGray    = "\033[0;37m"
	ansiMagenta = "\033[0;35m"
	ansiHeader  = "\033[48;5;153m\033[1;34m"

	homePageWidth      = 118
	homeRecentWidth    = 54
	homeStripGap       = 2
	homeCalloutWidth   = 36
	homeContentGap     = 2
	homePhotoColumns   = 4
	homePhotoColumnGap = 2
	homeFeaturedLimit  = 3

	homeEventsPlaceholder = "events data placeholder"
	homeSafetyNotice      = "Safety: If someone sends you a check, do not send them any money back. Never."
	homeAffiliationNotice = "SUpost is not sponsored by, endorsed by, or affiliated with Stanford University."
)

type styledWord struct {
	text  string
	color string
}

// RenderHomePosts renders the terminal homepage list.
func RenderHomePosts(w io.Writer, posts []domain.Post, featuredPosts []domain.Post, sections []domain.HomeCategorySection) error {
	now := time.Now()

	if err := RenderPageHeader(w, PageHeaderOptions{
		Width:      homePageWidth,
		Location:   "Stanford, California",
		RightLabel: "post",
		Now:        now,
	}); err != nil {
		return err
	}
	if err := renderHomePhotoStrip(w, posts, now, homePageWidth); err != nil {
		return err
	}

	if err := renderHomeOverviewAndRecent(w, posts, featuredPosts, sections, now, homePageWidth); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w); err != nil {
		return err
	}
	if err := RenderPageFooter(w, PageFooterOptions{Width: homePageWidth}); err != nil {
		return err
	}

	return nil
}

func renderHomeHeader(text string, width int) string {
	if width < len(text)+2 {
		return " " + text + " "
	}

	padding := width - len(text)
	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func renderHomePhotoStrip(w io.Writer, posts []domain.Post, now time.Time, width int) error {
	photos := selectRecentImagePosts(posts, homePhotoColumns)
	if len(photos) == 0 {
		return nil
	}

	calloutWidth, rightWidth := calculateStripWidths(width)
	columnWidth := photoColumnWidth(rightWidth, homePhotoColumns, homePhotoColumnGap)

	imageURLs := make([]string, 0, len(photos))
	titles := make([]string, 0, len(photos))
	timeAgo := make([]string, 0, len(photos))

	for _, post := range photos {
		imageURLs = append(imageURLs, formatTickerImageURL(post, now))
		titles = append(titles, strings.TrimSpace(post.Name))
		timeAgo = append(timeAgo, formatRelativeTime(postTimestamp(post), now))
	}

	rightRows := make([]string, 0, 8)
	rightRows = append(rightRows, renderWrappedColumnRows(imageURLs, columnWidth, "", homePhotoColumns)...)
	rightRows = append(rightRows, renderColumnRow(titles, columnWidth, ansiBlue, homePhotoColumns))
	rightRows = append(rightRows, renderColumnRow(timeAgo, columnWidth, ansiMagenta, homePhotoColumns))

	leftRows := renderHomeCalloutRows(calloutWidth)
	totalRows := len(rightRows)
	if len(leftRows) > totalRows {
		totalRows = len(leftRows)
	}

	for i := 0; i < totalRows; i++ {
		left := strings.Repeat(" ", calloutWidth)
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

func renderRecentPostRows(posts []domain.Post, now time.Time, wrapWidth int, sectionWidth int) []string {
	rows := make([]string, 0, len(posts)+1)
	rows = append(rows, ansiHeader+renderHomeHeader("recently posted", sectionWidth)+ansiReset)

	for _, post := range posts {
		title := formatPostTitle(post)
		email := formatDisplayEmail(post.Email)
		timeAgo := formatRelativeTime(postTimestamp(post), now)

		words := make([]styledWord, 0, 16)
		words = append(words, splitStyledWords(title, ansiBlue)...)
		words = append(words, splitStyledWords(email, ansiGray)...)
		if post.HasImage {
			words = append(words, styledWord{text: "📷"})
		}
		words = append(words, splitStyledWords(timeAgo, ansiMagenta)...)

		lines := wrapStyledWords(words, wrapWidth)
		for _, lineWords := range lines {
			rows = append(rows, renderStyledLine(lineWords))
		}
	}

	return rows
}

func renderHomeRecentAndFeaturedRows(posts []domain.Post, featuredPosts []domain.Post, now time.Time, sectionWidth int) []string {
	recentWidth, featuredWidth := splitHomeContentWidths(sectionWidth)
	if featuredWidth <= 0 {
		recentWrap := minInt(homeRecentWidth, recentWidth)
		return renderRecentPostRows(posts, now, recentWrap, recentWidth)
	}

	recentWrap := minInt(homeRecentWidth, recentWidth)
	recentRows := renderRecentPostRows(posts, now, recentWrap, recentWidth)
	featured := selectFeaturedJobPosts(featuredPosts, homeFeaturedLimit)
	if len(featured) == 0 {
		featured = selectFeaturedJobPosts(posts, homeFeaturedLimit)
	}
	featuredRows := renderFeaturedJobPostRows(featured, featuredWidth, featuredWidth)
	return combineHomeContentColumns(recentRows, featuredRows, recentWidth, featuredWidth)
}

func renderFeaturedJobPostRows(posts []domain.Post, wrapWidth int, sectionWidth int) []string {
	rows := make([]string, 0, len(posts)+1)
	rows = append(rows, ansiHeader+renderHomeHeader("featured job posts", sectionWidth)+ansiReset)

	for _, post := range posts {
		title := strings.TrimSpace(post.Name)
		if title == "" {
			title = "(untitled post)"
		}
		words := splitStyledWords(title, ansiBlue)
		if email := formatDisplayEmail(post.Email); email != "" {
			words = append(words, splitStyledWords(email, ansiGray)...)
		}
		lines := wrapStyledWords(words, wrapWidth)
		for _, lineWords := range lines {
			rows = append(rows, renderStyledLine(lineWords))
		}
	}

	rows = append(rows, strings.Repeat(" ", sectionWidth))
	rows = append(rows, ansiHeader+renderHomeHeader("events", sectionWidth)+ansiReset)
	appendWrappedTextRows(&rows, homeEventsPlaceholder, wrapWidth, ansiBlue)
	rows = append(rows, strings.Repeat(" ", sectionWidth))
	appendWrappedTextRows(&rows, homeSafetyNotice, wrapWidth, ansiBlue)
	appendWrappedTextRows(&rows, homeAffiliationNotice, wrapWidth, ansiGray)

	return rows
}

func appendWrappedTextRows(rows *[]string, text string, wrapWidth int, color string) {
	lines := wrapStyledWords(splitStyledWords(text, color), wrapWidth)
	for _, lineWords := range lines {
		*rows = append(*rows, renderStyledLine(lineWords))
	}
}

func selectFeaturedJobPosts(posts []domain.Post, limit int) []domain.Post {
	if limit <= 0 || len(posts) == 0 {
		return nil
	}

	filtered := make([]domain.Post, 0, limit)
	for _, post := range posts {
		if post.Status != domain.PostStatusActive {
			continue
		}
		if post.CategoryID != domain.CategoryJobsOffCampus {
			continue
		}
		filtered = append(filtered, post)
	}

	sort.Slice(filtered, func(i, j int) bool {
		iTime := featuredPostSortUnix(filtered[i])
		jTime := featuredPostSortUnix(filtered[j])
		if iTime == jTime {
			return filtered[i].ID > filtered[j].ID
		}
		return iTime > jTime
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}
	return filtered
}

func featuredPostSortUnix(post domain.Post) int64 {
	if ts := postTimestamp(post); !ts.IsZero() {
		return ts.Unix()
	}
	return 0
}

func splitHomeContentWidths(totalWidth int) (int, int) {
	if totalWidth <= 0 {
		return 0, 0
	}
	if totalWidth <= homeContentGap+2 {
		return totalWidth, 0
	}

	usable := totalWidth - homeContentGap
	left := usable / 2
	right := usable - left
	if left < 1 || right < 1 {
		return totalWidth, 0
	}
	return left, right
}

func combineHomeContentColumns(leftRows, rightRows []string, leftWidth, rightWidth int) []string {
	totalRows := len(leftRows)
	if len(rightRows) > totalRows {
		totalRows = len(rightRows)
	}

	gap := strings.Repeat(" ", homeContentGap)
	leftBlank := strings.Repeat(" ", leftWidth)
	rightBlank := strings.Repeat(" ", rightWidth)
	rows := make([]string, 0, totalRows)
	for i := 0; i < totalRows; i++ {
		left := leftBlank
		right := rightBlank
		if i < len(leftRows) {
			left = padANSIVisibleWidth(leftRows[i], leftWidth)
		}
		if i < len(rightRows) {
			right = padANSIVisibleWidth(rightRows[i], rightWidth)
		}
		rows = append(rows, left+gap+right)
	}
	return rows
}

func padANSIVisibleWidth(value string, width int) string {
	if width <= 0 {
		return ""
	}
	visible := ansiVisibleRuneLen(value)
	if visible >= width {
		return value
	}
	return value + strings.Repeat(" ", width-visible)
}

func ansiVisibleRuneLen(value string) int {
	length := 0
	inEscape := false
	for _, r := range value {
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
		length++
	}
	return length
}

func selectRecentImagePosts(posts []domain.Post, limit int) []domain.Post {
	if limit <= 0 {
		return nil
	}
	selected := make([]domain.Post, 0, limit)
	for _, post := range posts {
		if !post.HasImage {
			continue
		}
		selected = append(selected, post)
		if len(selected) == limit {
			break
		}
	}
	return selected
}

func formatTickerImageURL(post domain.Post, now time.Time) string {
	timestamp := post.TimePosted
	if timestamp <= 0 {
		if postedAt := postTimestamp(post); !postedAt.IsZero() {
			timestamp = postedAt.Unix()
		}
	}
	if timestamp <= 0 {
		timestamp = now.Unix()
	}
	return fmt.Sprintf("https://supost-prod.s3.amazonaws.com/posts/%d/ticker_%da?%d", post.ID, post.ID, timestamp)
}

func photoColumnWidth(totalWidth, columns, gap int) int {
	if columns <= 0 {
		return totalWidth
	}
	usable := totalWidth - ((columns - 1) * gap)
	if usable < columns {
		return 1
	}
	return usable / columns
}

func renderColumnRow(values []string, width int, color string, columns int) string {
	if columns <= 0 {
		columns = 1
	}
	cells := make([]string, 0, columns)
	for i := 0; i < columns; i++ {
		value := ""
		if i < len(values) {
			value = strings.TrimSpace(values[i])
		}
		cell := fitText(value, width)
		if color != "" && value != "" {
			cell = color + cell + ansiReset
		}
		cells = append(cells, cell)
	}
	return strings.Join(cells, "  ")
}

func renderWrappedColumnRows(values []string, width int, color string, columns int) []string {
	if columns <= 0 {
		columns = 1
	}

	columnLines := make([][]string, columns)
	maxLines := 1
	for i := 0; i < columns; i++ {
		value := ""
		if i < len(values) {
			value = strings.TrimSpace(values[i])
		}
		lines := wrapColumnValue(value, width)
		columnLines[i] = lines
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}

	rows := make([]string, 0, maxLines)
	for line := 0; line < maxLines; line++ {
		cells := make([]string, 0, columns)
		for col := 0; col < columns; col++ {
			segment := ""
			if line < len(columnLines[col]) {
				segment = columnLines[col][line]
			}
			cell := fitText(segment, width)
			if color != "" && segment != "" {
				cell = color + cell + ansiReset
			}
			cells = append(cells, cell)
		}
		rows = append(rows, strings.Join(cells, "  "))
	}
	return rows
}

func wrapColumnValue(value string, width int) []string {
	if width <= 0 {
		return []string{""}
	}

	runes := []rune(strings.TrimSpace(value))
	if len(runes) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, (len(runes)/width)+1)
	for len(runes) > width {
		lines = append(lines, string(runes[:width]))
		runes = runes[width:]
	}
	lines = append(lines, string(runes))
	return lines
}

func fitText(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) > width {
		if width == 1 {
			return "…"
		}
		return string(runes[:width-1]) + "…"
	}
	return value + strings.Repeat(" ", width-len(runes))
}

func formatPostTitle(post domain.Post) string {
	title := strings.TrimSpace(post.Name)
	if title == "" {
		title = "(untitled post)"
	}

	price := formatPrice(post.Price, post.HasPrice)
	if price == "" {
		return title
	}
	return title + " - " + price
}

func formatPrice(price float64, hasPrice bool) string {
	if !hasPrice {
		return ""
	}
	if price <= 0 {
		return "Free"
	}

	dollars := int64(math.Round(price))
	if math.Abs(price-float64(dollars)) < 0.001 {
		return "$" + formatIntWithCommas(dollars)
	}
	return fmt.Sprintf("$%.2f", price)
}

func formatIntWithCommas(v int64) string {
	s := strconv.FormatInt(v, 10)
	if len(s) <= 3 {
		return s
	}

	var out []byte
	prefix := len(s) % 3
	if prefix == 0 {
		prefix = 3
	}
	out = append(out, s[:prefix]...)
	for i := prefix; i < len(s); i += 3 {
		out = append(out, ',')
		out = append(out, s[i:i+3]...)
	}
	return string(out)
}

func postTimestamp(post domain.Post) time.Time {
	if !post.TimePostedAt.IsZero() {
		return post.TimePostedAt
	}
	if post.TimePosted > 0 {
		return time.Unix(post.TimePosted, 0)
	}
	if !post.CreatedAt.IsZero() {
		return post.CreatedAt
	}
	return time.Time{}
}

func formatRelativeTime(from, now time.Time) string {
	if from.IsZero() {
		return "about 0 minutes"
	}

	if from.After(now) {
		from = now
	}
	d := now.Sub(from)

	if d < time.Minute {
		return "about 1 minute"
	}
	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes <= 1 {
			return "about 1 minute"
		}
		return fmt.Sprintf("about %d minutes", minutes)
	}
	if d < 24*time.Hour {
		hours := int(d.Hours())
		if hours <= 1 {
			return "about 1 hour"
		}
		return fmt.Sprintf("about %d hours", hours)
	}

	days := int(d.Hours() / 24)
	if days <= 1 {
		return "about 1 day"
	}
	return fmt.Sprintf("about %d days", days)
}

func formatDisplayEmail(email string) string {
	normalized := strings.TrimSpace(email)
	if strings.Contains(strings.ToLower(normalized), "stanford.edu") {
		return "@stanford.edu"
	}
	return ""
}

func splitStyledWords(text, color string) []styledWord {
	fields := strings.Fields(strings.TrimSpace(text))
	words := make([]styledWord, 0, len(fields))
	for _, word := range fields {
		words = append(words, styledWord{text: word, color: color})
	}
	return words
}

func wrapStyledWords(words []styledWord, width int) [][]styledWord {
	if width <= 0 {
		return [][]styledWord{words}
	}
	if len(words) == 0 {
		return [][]styledWord{{}}
	}

	lines := make([][]styledWord, 0, len(words))
	current := make([]styledWord, 0, 8)
	currentWidth := 0

	for _, word := range words {
		wordLen := len([]rune(word.text))
		if wordLen > width {
			if len(current) > 0 {
				lines = append(lines, current)
				current = make([]styledWord, 0, 8)
				currentWidth = 0
			}

			runes := []rune(word.text)
			for len(runes) > width {
				chunk := string(runes[:width])
				lines = append(lines, []styledWord{{text: chunk, color: word.color}})
				runes = runes[width:]
			}
			if len(runes) > 0 {
				current = append(current, styledWord{text: string(runes), color: word.color})
				currentWidth = len(runes)
			}
			continue
		}

		needed := wordLen
		if len(current) > 0 {
			needed++
		}
		if currentWidth+needed <= width {
			current = append(current, word)
			currentWidth += needed
			continue
		}

		lines = append(lines, current)
		current = []styledWord{word}
		currentWidth = wordLen
	}

	if len(current) > 0 {
		lines = append(lines, current)
	}
	return lines
}

func renderStyledLine(words []styledWord) string {
	if len(words) == 0 {
		return ""
	}
	var b strings.Builder
	for i, word := range words {
		if i > 0 {
			b.WriteByte(' ')
		}
		if word.color != "" {
			b.WriteString(word.color)
			b.WriteString(word.text)
			b.WriteString(ansiReset)
		} else {
			b.WriteString(word.text)
		}
	}
	return b.String()
}
