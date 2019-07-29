# Contributing to Flux 

Make contributing to Flux easier. Share your Flux experience and feedback on technical hurdles you encounter.

Questions about Flux? Ask our [InfluxData Community](https://community.influxdata.com/)

## Report a bug

1. Search [existing issues](https://github.com/influxdata/flux/issues) to verify the bug hasn’t been reported.
2. Open a [new issue](https://github.com/influxdata/flux/issues/new), and include these details:
   * Operating system (or distribution), for example, 64-bit Ubuntu 14.04
   * Flux version
   * Whether Flux was installed from a package or built from source
   * Test cases, if applicable, to demonstrate the issue
   
        **_Tip_**: The easier it is for us to reproduce the bug, the faster we can fix it. For handy bug reporting tips, check out [Simon Tatham's essay "How to Report Bugs Effectively."](http://www.chiark.greenend.org.uk/~sgtatham/bugs.html)

## Request a feature

* To request a new feature, open a [new issue](https://github.com/influxdata/flux/issues/new). Your feature requests help us prioritize work and improve Flux. We appreciate your detailed requirements!

## Sign our Contributor License Agreement (CLA)

Before submitting a pull request to contribute to Flux, you must sign the [InfluxData Contributor License Agreement](https://www.influxdata.com/legal/cla/) (CLA).

## Third-party packages

We prefer writing a little code over using a third-party package (one that’s not in the standard Go distribution). To maximize your change being accepted, use standard Go libraries or third-party packages that we've included with Flux.

For rationale, check out the post [The Case Against Third Party Libraries](http://blog.gopheracademy.com/advent-2014/case-against-3pl/).

## Add a function

1. To add a function, see required guidelines for:
   * [Flux functions](https://github.com/influxdata/flux/docs/contributing/Flux_Functions.md)
   * [Scalar functions](https://github.com/influxdata/flux/docs/contributing/Scalar_Functions.md)
   * [Source and sink functions](https://github.com/influxdata/flux/docs/contributing/Source_Sink_Functions.md)
   * [Stream transformation functions](https://github.com/influxdata/flux/docs/contributing/Stream_Transformation_Functions.md)

2. (Optional) Open a [new issue](https://github.com/influxdata/flux/issues/new) to discuss the changes you would like to make before creating a function. This often helps reduce rework.

3. After completing the required guidelines to add the function, submit a new pull request.

## Submit a pull request

1. Fork the Flux repository.
2. Create a feature branch of your fork, and then add your change.
3. Run the test suite with your change. (Changes that cause tests to fail cannot be merged.)
4. Commit your change. (Use the Flux conventional commit message format: https://www.conventionalcommits.org/en/v1.0.0-beta.3/.)
5. Create a pull request from your feature branch against master of the Flux repository. In your pull request, include details of your change and test cases, if applicable.

## Useful links

- [Go in production](http://peter.bourgon.org/go-in-production/)
- [Principles of designing Go APIs with channels](https://inconshreveable.com/07-08-2014/principles-of-designing-go-apis-with-channels/)
- [Common mistakes in Golang](http://soryy.com/blog/2014/common-mistakes-with-go-lang/) ()
  especially the _Loops, Closures, and Local Variables_ section)
