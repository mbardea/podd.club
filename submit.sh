#!/bin/bash

curl -X POST -d "url=$1" http://localhost:3000/api/users/1/categories/1/schedule-download 

