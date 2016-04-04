unigornel
=========

[![Build Status](https://jenkins.unigornel.org/buildStatus/icon?job=unigornel-master)](https://jenkins.unigornel.org/job/unigornel-master/)

To clone this repository use

```
$ git clone --recursive git@github.ugent.be:unigornel/unigornel
```

or

```
$ git clone git@github.ugent.be:unigornel/unigornel
$ git submodule update --init --recursive
```

## Dependencies

```
# apt-get install time
# apt-get install python3 python3-pip
# pip3 install junit-xml
```

## Scripts

  - `build.bash`: build the Go port, minios, or a unigornel application
  - `test.bash`: run test suite (use `--no-go` to skip building Go)
  - `integration_tests/test.py`: integration tests runner
