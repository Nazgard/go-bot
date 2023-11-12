package telegram

import (
	"fmt"
	"github.com/nleeper/goment"
	"makarov.dev/bot/internal/config"
	"strings"
	"time"
)

var location, _ = time.LoadLocation("Europe/Moscow")
var beautifulDay, _ = goment.New(goment.DateTime{
	Year:     2019,
	Month:    int(time.April),
	Day:      5,
	Hour:     19,
	Minute:   30,
	Location: location})

func init() {
	err := AddRouterFunc("/dd", ddCmd)
	if err != nil {
		config.GetLogger().Errorf("Error while add telegram DD cmd %s", err.Error())
		return
	}
}

func ddCmd(txt string) string {
	var from goment.Goment
	var to goment.Goment
	if txt == "" {
		from = *beautifulDay
		to1, _ := goment.New()
		to = *to1
	}
	if txt != "" {
		split := strings.Split(txt, " ")
		if len(split) == 1 {
			parse, err := time.ParseInLocation(dateParseLayout, split[0], location)
			if err != nil {
				return err.Error()
			}
			from1, _ := goment.New(parse)
			to1, _ := goment.New()
			from = *from1
			to = *to1
		}
		if len(split) == 2 {
			parse1, err := time.Parse(dateParseLayout, split[0])
			if err != nil {
				return err.Error()
			}
			parse2, err := time.Parse(dateParseLayout, split[1])
			if err != nil {
				return err.Error()
			}
			from1, _ := goment.New(parse1)
			to1, _ := goment.New(parse2)
			from = *from1
			to = *to1
		}
	}

	rawCount := duration(from.ToTime(), to.ToTime())

	monthCount := to.Diff(from, "months")
	dayCount := to.Diff(from, "days")
	if monthCount != 0 {
		yearCount := float32(dayCount) / 365.0
		return fmt.Sprintf("%s (~%.2f года)", rawCount, yearCount)
	}

	return rawCount
}

func duration(a, b time.Time) string {
	d := b.Sub(a)

	if d < 0 {
		d *= -1
	}

	if d < day {
		return d.String()
	}

	n := d / day
	d -= n * day

	if d == 0 {
		return fmt.Sprintf("%dd", n)
	}

	return fmt.Sprintf("%dd%s", n, d)
}
