// Package bigtable provides tools for working with data in
// [Google Cloud Bigtable](https://cloud.google.com/bigtable/) databases.
//
// ## Metadata
// introduced: 0.45.0
//
package bigtable


// from retrieves data from a [Google Cloud Bigtable](https://cloud.google.com/bigtable/) data source.
//
// ## Parameters
//
// - token: Google Cloud IAM token to use to access the Cloud Bigtable database.
//
//   For more information, see the following:
//
//   - [Cloud Bigtable Access Control](https://cloud.google.com/bigtable/docs/access-control)
//   - [Google Cloud IAM How-to guides](https://cloud.google.com/iam/docs/how-to)
//   - [Setting Up Authentication for Server to Server Production Applications on Google Cloud](https://cloud.google.com/docs/authentication/production)
//
// - project: Cloud Bigtable project ID.
// - instance: Cloud Bigtable instance ID.
// - table: Cloud Bigtable table name.
//
// ## Examples
// ### Query Google Cloud Bigtable
// ```no_run
// import "experimental/bigtable"
//
// bigtable.from(
//     token: "example-token",
//     project: "example-project",
//     instance: "example-instance",
//     table: "example-table",
// )
// ```
//
// ## Metadata
// tags: inputs
//
builtin from : (token: string, project: string, instance: string, table: string) => stream[T]
    where
    T: Record
