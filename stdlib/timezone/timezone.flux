// Package timezone defines functions for setting timezones
// on the location option in package universe.
//
// introduced: 0.134.0
//
package timezone


// utc is the default location with a completely linear clock and no offset.
// It is used as the default for location-related options.
utc = {zone: "UTC", offset: 0h}

// fixed returns a location record with a fixed offset.
//
// ## Parameters
// - offset: Fixed duration for the location offset.
//   This duration is the offset from UTC.
//
// ## Examples
//
// ### Return a fixed location record
// ```no_run
// import "timezone"
//
// timezone.fixed(offset: -8h)
//
// // Returns {offset: -8h, zone: "UTC"}
// ```
//
// ### Set the location option using a fixed location
// ```
// import "timezone"
//
// // This results in midnight at 00:00:00-08:00 on any day.
// option location = timezone.fixed(offset: -8h)
// ```
//
// tags: date/time,location
//
fixed = (offset) => ({zone: utc.zone, offset: offset})

// location returns a location record based on a location or timezone name.
//
// ## Parameters
// - `name` is the name of the location as defined by the tzdata database.
//
// ## Examples
//
// ### Return a timezone-based location record
// ```no_run
// import "timezone"
//
// timezone.location(name: "America/Los_Angeles")
//
// // Returns {offset: 0ns, zone: "America/Los_Angeles"}
// ```
//
// ### Set the location option using a timezone-based location
// ```
// import "timezone"
//
// option location = timezone.location(name: "America/Los_Angeles")
// ```
//
// tags: date/time,location
//
builtin location : (name: string) => {zone: string, offset: duration}
