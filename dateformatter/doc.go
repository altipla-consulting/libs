// Package dateformatter helps formatting date in the correct locale. It uses
// data extracted from CLDR v27.0.1.
//
// It has the same formatting options as the time package of the standard library,
// excluding timezones and nanoseconds as they are not needed in a daily basis
// to format user-facing dates and times.
//
// It has support for the following langs: spanish, english, french and russian.
// More can be added on demand modifying the generator to create the symbols files.
package dateformatter
