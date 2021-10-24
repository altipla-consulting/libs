package api

// DateTimeFormat to send & receive from RavenDB. Nanoseconds are important or
// comparisons and index operations will fail in mysterious ways.
// Zeros for the nanoseconds is bad on purpose to avoid formatting them.
const DateTimeFormat = "2006-01-02T15:04:05.0000000Z"
