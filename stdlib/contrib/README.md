# Contribute Third Party Flux Packages

This directory contains a collections of third party contributions to Flux.
This is a simple and light weight process for sharing Flux code with the wider community while no other official means of sharing Flux packages exists.

## Adding a New Package

To add a new package first create a new directory using your github account name.
Then create another directory for the name of your package.

For example if I (@nathanielc) wanted to add a Flux package about reading text based ledger files in Flux I would create a directory named `nathanielc/ledger` under the `contrib` directory.
I would then place all the Go code, Flux code and corresponding test cases into that directory.

Please see the [CONTRIBUTING](https://github.com/influxdata/flux/blob/master/CONTRIBUTING.md) guide for more details on how to make contributions to the Flux repo.

## Package Ownership

Packages in the `contrib` directory are owned and maintained by their author not the InfluxData team.
As such the author will be requested for review on all changes to the package.


## Future Plans

In the future we may create a more official repository of Flux packages that does not require committing the code to the Flux code repository.
When that happens we intend to promote the packages from the `contrib` directory into their own Flux package in that repository in what ever form that takes.
Until then we will collect third party packages into the `contrib` directory.

