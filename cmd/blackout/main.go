package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
)

const (
	offCellColor     = "#ff3333"
	defaultCellColor = "#ffffff"
	queueNumber      = "1.2"
	targetPidcherga  = "2"
	url              = "https://www.ztoe.com.ua/unhooking-search.php?pidcherga_id=" + targetPidcherga
)

// WaybarOutput matches the Waybar custom module JSON schema
type WaybarOutput struct {
	Text    string `json:"text"`
	Alt     string `json:"alt"`
	Tooltip string `json:"tooltip"`
	Class   string `json:"class"`
}

// GenerateTimeIntervals generates forty-eight 30-minute time intervals for 24 hours
func GenerateTimeIntervals() []string {
	intervals := make([]string, 0, 48)
	for i := 0; i < 48; i++ {
		startTotalMinutes := i * 30
		endTotalMinutes := (i + 1) * 30

		startHour := startTotalMinutes / 60
		startMin := startTotalMinutes % 60

		endHour := endTotalMinutes / 60
		endMin := endTotalMinutes % 60

		interval := fmt.Sprintf("%02d:%02d-%02d:%02d", startHour, startMin, endHour, endMin)
		intervals = append(intervals, interval)
	}
	return intervals
}

// setupCellParser sets up the Colly parser to extract color statuses from table cells
func setupCellParser(c *colly.Collector, targetPidchergaId string, offColor string, defaultColor string, colorsArr *[]string) error {
	targetHref := "/unhooking-search.php?pidcherga_id=" + targetPidchergaId

	c.OnHTML("td:has(a)", func(e *colly.HTMLElement) {
		href := e.ChildAttr("a", "href")
		if href != targetHref {
			return
		}

		cellStyle := e.Attr("style")
		colorStatus := ""

		if strings.Contains(cellStyle, offColor) {
			colorStatus = "Червона"
		} else if strings.Contains(cellStyle, defaultColor) {
			colorStatus = "Біла"
		}

		if colorStatus != "" {
			*colorsArr = append(*colorsArr, colorStatus)
		}
	})

	return c.Visit(url)
}

// GenerateBlackoutLogic parses time slots, groups outage intervals, and checks current power status
func GenerateBlackoutLogic(c *colly.Collector, now time.Time) ([]string, bool) {
	ongoingOutage := false
	isCurrentlyOff := false
	startTime := ""
	endTime := ""
	result := []string{}

	// Calculate index for current 30-minute slot (0..47)
	currentIndex := now.Hour() * 2
	if now.Minute() >= 30 {
		currentIndex++
	}

	interval := GenerateTimeIntervals()
	colorsArr := make([]string, 0, 48)

	err := setupCellParser(c, targetPidcherga, offCellColor, defaultCellColor, &colorsArr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fetch error: %v\n", err)
		return result, false
	}

	for i, val := range colorsArr {
		// Check current status against calculated time slot
		if i == currentIndex && val == "Червона" {
			isCurrentlyOff = true
		}

		parts := strings.Split(interval[i], "-")
		if val == "Червона" {
			if !ongoingOutage {
				ongoingOutage = true
				startTime = parts[0]
			}
			endTime = parts[1]

		} else if val == "Біла" {
			if ongoingOutage {
				results := fmt.Sprintf("• %s - %s", startTime, endTime)
				result = append(result, results)
				ongoingOutage = false
			}
		}
	}

	// Handle edge case where outage extends to the end of the day (24:00)
	if ongoingOutage {
		results := fmt.Sprintf("• %s - %s", startTime, endTime)
		result = append(result, results)
	}

	return result, isCurrentlyOff
}

// PrintWaybarOutput builds and outputs the JSON response for Waybar
func PrintWaybarOutput(tooltipLines []string, isCurrentlyOff bool, timeNow time.Time) error {
	header := fmt.Sprintf("Графік відключень (Черга %s):", queueNumber)
	formattedTooltip := []string{header}

	if len(tooltipLines) == 0 {
		formattedTooltip = append(formattedTooltip, "Відключень не заплановано")
	} else {
		formattedTooltip = append(formattedTooltip, tooltipLines...)
	}

	updateTimeStr := fmt.Sprintf("Оновлено о: %s", timeNow.Format("15:04"))
	formattedTooltip = append(formattedTooltip, "", updateTimeStr)

	output := WaybarOutput{
		Tooltip: strings.Join(formattedTooltip, "\n"),
	}

	if isCurrentlyOff {
		output.Text = "Немає світла"
		output.Alt = "off"
		output.Class = "off"
	} else {
		output.Text = "Світло є"
		output.Alt = "on"
		output.Class = "on"
	}

	jsonData, err := json.Marshal(output)
	if err != nil {
		return fmt.Errorf("JSON encoding error: %w", err)
	}

	fmt.Println(string(jsonData))
	return nil
}

func main() {
	c := colly.NewCollector()
	now := time.Now()

	// Parse blackout schedules and evaluate current status
	blackoutIntervals, isOff := GenerateBlackoutLogic(c, now)

	// Print formatted JSON payload for Waybar
	err := PrintWaybarOutput(blackoutIntervals, isOff, now)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Waybar output error: %v\n", err)
		os.Exit(1)
	}
}
