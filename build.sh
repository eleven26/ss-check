#!/bin/bash

go build ./... && tar -zcf ss-check_${1}.tar.gz ss-check && rm ss-check
