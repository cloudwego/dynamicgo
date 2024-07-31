#! /bin/bash


# please install fastcodec version of thriftgo
# if the `feat-fastcodec` branch doesn't exist try to use `main` branch
# the feature is only for internal use now, but it will be released in the future when it's stable.
# go install github.com/cloudwego/thriftgo@feat-fastcodec


IDL_LIST=(
"./idl/base.thrift"
"./idl/baseline.thrift"
"./idl/example.thrift"
"./idl/example2.thrift"
"./idl/example3.thrift"
#"./idl/example4.thrift" # only used by thrift.NewDescritorFromPath
"./idl/null.thrift"
"./idl/ref.thrift"
"./idl/skip.thrift"
"./idl/deep/deep.ref.thrift"
)
THRIFTGO_ARGS="-g fastgo:no_default_serdes=true,package_prefix=github.com/cloudwego/dynamicgo/testdata/kitex_gen -o=./kitex_gen"

for idl in "${IDL_LIST[@]}"
do
  echo "generating" $idl
  thriftgo $THRIFTGO_ARGS $idl
done
