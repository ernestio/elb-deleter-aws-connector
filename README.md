# ELB deleter aws connector

Service to delete aws ELB's, it responds to *elb.delete.aws* and will respond with *elb.delete.aws.done* or *elb.delete.aws.error*

## Build status

* master: [![CircleCI](https://circleci.com/gh/ernestio/elb-deleter-aws-connector/tree/master.svg?style=svg)](https://circleci.com/gh/ernestio/elb-deleter-aws-connector/tree/master)
* develop: [![CircleCI](https://circleci.com/gh/ernestio/elb-deleter-aws-connector/tree/develop.svg?style=svg)](https://circleci.com/gh/ernestio/elb-deleter-aws-connector/tree/develop)

## Installation

```
make deps
make install
```

## Running Tests

```
make test
```

## Contributing

Please read through our
[contributing guidelines](CONTRIBUTING.md).
Included are directions for opening issues, coding standards, and notes on
development.

Moreover, if your pull request contains patches or features, you must include
relevant unit tests.

## Versioning

For transparency into our release cycle and in striving to maintain backward
compatibility, this project is maintained under [the Semantic Versioning guidelines](http://semver.org/).

## Copyright and License

Code and documentation copyright since 2015 r3labs.io authors.

Code released under
[the Mozilla Public License Version 2.0](LICENSE).
