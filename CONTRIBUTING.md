# Contributing to Flux

## Bug reports
Before you file an issue, please search existing issues in case it has already
been filed, or perhaps even fixed. If you file an issue, please include the following.

* Full details of your operating system (or distribution) e.g. 64-bit Ubuntu 14.04.
* The version of Flux you are running.
* Whether you installed it using a pre-built package, or built it from source.
* A small test case, if applicable, that demonstrates the issues.

Remember the golden rule of bug reports: **The easier you make it for us to reproduce
the problem, the faster it will get fixed.** If you have never written a bug report
before, or if you want to brush up on your bug reporting skills, we recommend reading
[Simon Tatham's essay "How to Report Bugs Effectively."](http://www.chiark.greenend.org.uk/~sgtatham/bugs.html)

Please note that issues are *not the place to file general questions* such as
"how do I use InfluxDB with Flux?" Questions of this nature should be sent to the
[InfluxData Community](https://community.influxdata.com/), not filed as issues.
Issues like this will be closed.

## Feature requests
We really like to receive feature requests as they help us prioritize our work.
Please be clear about your requirements. Incomplete feature requests may simply
be closed if we don't understand what you would like to see added to Flux.

## Submitting a pull request
To submit a pull request you should fork the Flux repository and make your change
on a feature branch of your fork. Then generate a pull request from your branch
against **master** of the Flux repository. Include in your pull request details of
your change -- the **why** *and* the **how** -- as well as the testing you performed.
Also, be sure to run the test suite with your change in place. Changes that cause
tests to fail cannot be merged.

There will usually be some back and forth as we finalize the change, but once
that completes, it may be merged.

To assist in review for the PR, please add the following to your pull request comment:

```md
- [ ] Sign [CLA](https://www.influxdata.com/legal/cla/) (if not already signed)
```

Flux uses _conventional commit message_ formats: https://www.conventionalcommits.org/en/v1.0.0-beta.3/. Please use this commit message format for commits that will be visible in influxdata/flux history.

## Signing the CLA
In order to contribute back to Flux, you must sign the
[InfluxData Contributor License Agreement](https://www.influxdata.com/legal/cla/) (CLA).

## Use of third-party packages
A third-party package is defined as one that is not part of the standard Go distribution.
Generally speaking, we prefer to minimize our use of third-party packages and avoid
them unless absolutely necessarily. We'll often write a little bit of code rather
than pull in a third-party package. To maximize the chance your change will be accepted,
use only the standard libraries, or the third-party packages we have decided to use.

For rationale, check out the post [The Case Against Third Party Libraries](http://blog.gopheracademy.com/advent-2014/case-against-3pl/).

## Useful links
- [Useful techniques in Go](http://arslan.io/ten-useful-techniques-in-go)
- [Go in production](http://peter.bourgon.org/go-in-production/)
- [Principles of designing Go APIs with channels](https://inconshreveable.com/07-08-2014/principles-of-designing-go-apis-with-channels/)
- [Common mistakes in Golang](http://soryy.com/blog/2014/common-mistakes-with-go-lang/).
  Especially this section `Loops, Closures, and Local Variables`
