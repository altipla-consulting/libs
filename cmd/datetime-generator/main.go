// Package generator helps with creating the symbols files for supported langs
// extracting the data from CLDR.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/text/unicode/cldr"

	"libs.altipla.consulting/errors"
)

var locales = flag.String("locales", "", "Locales to extract from CLDR")

var weekdaysOrder = map[string]int{
	"sun": 0,
	"mon": 1,
	"tue": 2,
	"wed": 3,
	"thu": 4,
	"fri": 5,
	"sat": 6,
}

func main() {
	flag.Parse()

	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	log.Println("Extract CLDR data")

	coreFile, err := os.Open("/tmp/core.zip")
	if err != nil {
		return errors.Trace(err)
	}

	decoder := cldr.Decoder{}
	decoder.SetDirFilter("main")
	decoder.SetSectionFilter("dates")
	data, err := decoder.DecodeZip(coreFile)
	if err != nil {
		return errors.Trace(err)
	}

	for _, locale := range strings.Split(*locales, ",") {
		log.Println("Extract locale", locale)

		ldml, err := data.LDML(locale)
		if err != nil {
			return errors.Trace(err)
		}

		if err := extractLDML(locale, ldml); err != nil {
			return errors.Trace(err)
		}
	}

	log.Println("Generator done")

	return nil
}

func mustWriteln(dest io.Writer, args ...interface{}) {
	if _, err := fmt.Fprintln(dest, args...); err != nil {
		panic(err)
	}
}

func mustWritef(dest io.Writer, format string, args ...interface{}) {
	if _, err := fmt.Fprintf(dest, format, args...); err != nil {
		panic(err)
	}
}

func extractLDML(locale string, ldml *cldr.LDML) error {
	var blocks []string
	for _, calendar := range ldml.Dates.Calendars.Calendar {
		if calendar.Type != "gregorian" {
			continue
		}

		for _, ctx := range calendar.Months.MonthContext {
			if ctx.Type != "format" {
				continue
			}

			for _, width := range ctx.MonthWidth {
				if width.Type != "wide" && width.Type != "abbreviated" {
					continue
				}

				months := make([]string, 13)
				months[0] = "---"
				for _, month := range width.Month {
					n, err := strconv.ParseInt(month.Type, 10, 64)
					if err != nil {
						return errors.Trace(err)
					}

					months[n] = month.Data()
				}

				nameType := "LongMonthNames"
				if width.Type == "abbreviated" {
					nameType = "ShortMonthNames"
				}

				var buf bytes.Buffer
				mustWritef(&buf, "\t%s[`%s`] = []string{\n", nameType, locale)
				for _, m := range months {
					mustWritef(&buf, "\t\t`%s`,\n", m)
				}
				mustWritef(&buf, "\t}")
				blocks = append(blocks, buf.String())
			}
		}

		for _, ctx := range calendar.Days.DayContext {
			if ctx.Type != "format" {
				continue
			}

			for _, width := range ctx.DayWidth {
				if width.Type != "wide" && width.Type != "abbreviated" {
					continue
				}

				weekdays := make([]string, 7)
				for _, day := range width.Day {
					weekdays[weekdaysOrder[day.Type]] = day.Data()
				}

				nameType := "LongWeekdays"
				if width.Type == "abbreviated" {
					nameType = "ShortWeekdays"
				}

				var buf bytes.Buffer
				mustWritef(&buf, "\t%s[`%s`] = []string{\n", nameType, locale)
				for _, d := range weekdays {
					mustWritef(&buf, "\t\t`%s`,\n", d)
				}
				mustWritef(&buf, "\t}")
				blocks = append(blocks, buf.String())
			}
		}
	}

	dest, err := os.Create(fmt.Sprintf("datetime/symbols/%s.go", locale))
	if err != nil {
		return errors.Trace(err)
	}
	defer dest.Close()

	mustWriteln(dest, "// GENERATED FROM CLDR DATA. DO NOT EDIT MANUALLY.")
	mustWriteln(dest)
	mustWriteln(dest, "package symbols")
	mustWriteln(dest)
	mustWriteln(dest, "func init() {")

	for i, block := range blocks {
		if i > 0 {
			mustWriteln(dest)
		}
		mustWriteln(dest, block)
	}

	mustWriteln(dest, "}")
	mustWriteln(dest)

	return nil
}
