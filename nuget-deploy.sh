#!/bin/bash

which dotnet > /dev/null

if [ $? -ne 0 ]; then
	echo "dotnet not available."
	exit
fi;

