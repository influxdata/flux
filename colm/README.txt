#!/bin/bash
#

#
# This directory contains some colm programs that can be used standalone, or in
# the case of tableflux, embedded into the Go codebase.
#
#  * parseflux: contains a plain parsing driver for pure flux
#
#  * influxql: contains timeboxed experiment to translate influxql -> flux. Very
#    far from complete.
#
#  * tableflux: contains a translator from TableFlux (proposed early 2020) to Flux.
#    This code can be included in the Go codebase.
#
# In the case of tableflux, by default, a stubbed interface will be included in
# the Go project. To make it functional you must first install colm, then
# configure and build in this directory with the go interface enabled. It will
# generate the appropriate Go/C files. See tableflux/call.*.in. These are
# rewritten by make.
#
# If you have trouble building, the first thing to do is get the latest colm on
# the master branch. New features may be added to colm to support the work
# here.
#

ORIG_PWD=$PWD

#
# 1. Install colm
#

BUILD=$HOME/build
INST=$HOME/pkgs
VERSION=0.14.1

mkdir -p $BUILD $INST

cd $BUILD
wget https://www.colm.net/files/colm/colm-$VERSION.tar.gz
tar -xzf colm-$VERSION.tar.gz
cd colm-$VERSION

./configure --prefix=$INST/colm-$VERSION --disable-manual
make install

#
# 2. Configure in this directory, giving location of colm install, then make
#

cd $ORIG_PWD
./configure --with-colm=$INST/colm-$VERSION --enable-go-interface
make

exit

#
# After this you should be able to build flux or influxdb (with go mod rewrite
# to this flux repos) and it should pick up the TableFlux implementation.
#
# To disable the implementation after enabling it, reconfigure without the
# --enable-go-interface flag and rebuild.
#

#
# To use the standalone tableflux program:
#
cd tableflux

# set ORG, TOKEN and INFLUX (full path to command)
vim .fluxrc

# make a bucket called tableflux
./run create tableflux

# write the example data
./run write tableflux data3.pts

# try some queries
./run tableflux query17.flux

#
# To just see the output of the transformation from tableflux to flux:
#
./tableflux < tableflux.flux
