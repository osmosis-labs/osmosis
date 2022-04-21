# this script is for generating protobuf files for the new google.golang.org/protobuf API



set -eo pipefail

protoc_install_gocosmos() {
   echo "Installing protobuf gocosmos plugin"
   # we should use go install, but regen-network/cosmos-proto contains
   # replace directives. It must not contain directives that would cause
   # it to be interpreted differently than if it were the main module.
   # So the command below issues a warning and we are muting it for now.
   #
   # Installing plugins must be done outside of the module
   (go get github.com/regen-network/cosmos-proto/protoc-gen-gocosmos@v0.3.1 2> /dev/null)
 }

protoc_install_gocosmos

echo "Generating gogo proto code"
cd proto
proto_dirs=$(find ./osmosis -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  for file in $(find "${dir}" -maxdepth 1 -name '*.proto'); do
    if grep "option go_package" $file &> /dev/null ; then
      buf generate --template buf.gen.gogo.yaml $file
    fi
  done
done

cd ..

# generate codec/testdata proto code
#(cd testutil/testdata; buf generate)

# move proto files to the right places
cp -r github.com/osmosis-labs/osmosis/v7/* ./
rm -rf github.com

protoc_gen_install() {
  go install github.com/cosmos/cosmos-proto/cmd/protoc-gen-go-pulsar@latest #2>/dev/null
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest #2>/dev/null
  go install github.com/cosmos/cosmos-sdk/orm/cmd/protoc-gen-go-cosmos-orm@latest #2>/dev/null
}

protoc_gen_install

echo "Generating API module"
#(cd api; find ./ -type f \( -iname \*.pulsar.go -o -iname \*.pb.go -o -iname \*.cosmos_orm.go -o -iname \*.pb.gw.go \) -delete; find . -empty -type d -delete; cd ..)
(cd proto; buf generate --template buf.gen.pulsar.yaml)