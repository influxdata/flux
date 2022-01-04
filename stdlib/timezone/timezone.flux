// Package timezone defines functions for setting timezones
// on the location option in package location.
package timezone


// utc is the default location with a completely linear clock
// and no offset. It is used as the default for location related
// options.
utc = {zone: "UTC", offset: 0h}

// fixed is a function that constructs a location with a fixed offset.
//
// ## Parameters
// - `offset` is the fixed duration for the location offset.
//   This duration is the offset from UTC.
//
// ## Example
//
// ```
// import "timezone"
//
// // This results in midnight at 00:00:00-08:00 on any day.
// option location = timezone.fixed(offset: -8h)
//
// from(...)
//   |> range(...)
//   |> window(every: 1d)
fixed = (offset) => ({zone: utc.zone, offset: offset})

// location loads a timezone based on a location name.
//
// ## Parameters
// - `name` is the name of the location as defined by the tzdata database.
//
// ## Example
//
// ```
// import "timezone"
//
// option location = timezone.location(name: "America/Los_Angeles")
//
// from(...)
//   |> range(...)
//   |> window(every: 1d)
builtin location : (name: string) => {zone: string, offset: duration}
